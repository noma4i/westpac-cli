package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

type AuthResult struct {
	ProfileID   string
	AccessToken string
	XSRFToken   string
}

type eamResponse struct {
	Reference eamReference   `json:"reference"`
	Operations []eamOperation `json:"operations"`
	Keymap     eamKeymap      `json:"keymap"`
}

type eamReference struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

type eamOperation struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	SubmitToUri string            `json:"submitToUri"`
	Input       map[string]string `json:"input"`
}

type eamKeymap struct {
	Keys  []map[string]int `json:"keys"`
	Halgm string           `json:"halgm"`
	Malgm string           `json:"malgm"`
}

// obfuscatePassword converts password using the EAM keymap.
// Each char is mapped via keymap to an index, then malgm[index] is the replacement.
func obfuscatePassword(password string, keymap eamKeymap) string {
	charMap := make(map[rune]int)
	for _, entry := range keymap.Keys {
		for ch, num := range entry {
			if len(ch) > 0 {
				charMap[rune(ch[0])] = num
			}
		}
	}

	malgm := []rune(keymap.Malgm)
	var result []rune
	for _, ch := range password {
		if idx, ok := charMap[ch]; ok && idx < len(malgm) {
			result = append(result, malgm[idx])
		}
	}
	return string(result)
}

// Login performs the full authentication flow:
// 1. getEamInterfaceInfo - get keymap for password obfuscation
// 2. AuthenticateHttpServlet - authenticate with obfuscated password
// 3. OAuth session + PKCE authorize
// 4. Token exchange
// 5. Select profile
func (c *Client) Login(customerID, password string) (*AuthResult, error) {
	// Step 1: Get EAM Interface Info
	c.log("[AUTH] Step 1: getEamInterfaceInfo")
	status, body, err := c.doGet(fmt.Sprintf("/eam/servlet/getEamInterfaceInfo?uid=%s", customerID), nil)
	if err != nil {
		return nil, fmt.Errorf("step 1: %w", err)
	}
	c.log(fmt.Sprintf("[AUTH] Step 1: status=%d", status))

	var eam eamResponse
	if err := json.Unmarshal(body, &eam); err != nil {
		return nil, fmt.Errorf("step 1: parsing EAM: %w", err)
	}
	if len(eam.Keymap.Keys) == 0 || eam.Keymap.Malgm == "" {
		return nil, fmt.Errorf("step 1: no keymap in response")
	}

	// Step 2: Authenticate with obfuscated password
	c.log("[AUTH] Step 2: AuthenticateHttpServlet")
	obfPwd := obfuscatePassword(password, eam.Keymap)

	authForm := url.Values{
		"username": {customerID},
		"brand":    {BrandSilo},
		"halgm":   {eam.Keymap.Halgm},
		"password": {obfPwd},
	}

	// Direct HTTP call - 302 is expected success, don't treat HTML as error
	authReq, err := http.NewRequest("POST", c.baseURL+"/eam/servlet/AuthenticateHttpServlet",
		strings.NewReader(authForm.Encode()))
	if err != nil {
		return nil, fmt.Errorf("step 2: creating request: %w", err)
	}
	authReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	authResp, err := c.http.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("step 2: %w", err)
	}
	authRespBody, _ := io.ReadAll(authResp.Body)
	authResp.Body.Close()

	authLocation := authResp.Header.Get("Location")
	c.log(fmt.Sprintf("[AUTH] Step 2: status=%d, location=%s", authResp.StatusCode, authLocation))

	// 302 is expected - check where it redirects
	if authResp.StatusCode == 302 {
		if strings.Contains(authLocation, "TAM_OP=error") {
			return nil, fmt.Errorf("authentication failed: invalid credentials")
		}
		// Follow the redirect chain to collect session cookies
		if authLocation != "" {
			c.log("[AUTH] Step 2: following redirect chain")
			err = c.followRedirects(authLocation)
			if err != nil {
				c.log(fmt.Sprintf("[AUTH] Step 2: redirect error: %v", err))
			}
		}
	} else if authResp.StatusCode >= 400 {
		return nil, fmt.Errorf("authentication failed (status %d): %s", authResp.StatusCode, truncate(string(authRespBody), 200))
	}

	// Step 3: OAuth session
	c.log("[AUTH] Step 3: OAuth session")
	sessionBody := map[string]string{
		"auth_method":   "password",
		"client_id":     "wlivemobile",
		"scope":         "SSI",
		"client_secret": c.deviceID,
		"grant_type":    "password",
	}
	status, body, err = c.doJSON("POST", "/eam/servlet/oauth/oauth20/v2/session?scope=WLL10", sessionBody, nil)
	if err != nil {
		c.log(fmt.Sprintf("[AUTH] Step 3 error: %v", err))
	} else {
		c.log(fmt.Sprintf("[AUTH] Step 3: status=%d, body=%s", status, truncate(string(body), 500)))
	}

	// Step 4: PKCE Authorization
	c.log("[AUTH] Step 4: PKCE authorize")
	verifier, challenge, err := utils.GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("generating PKCE: %w", err)
	}
	nonce, _ := utils.RandomString(16)
	state, _ := utils.RandomString(16)

	authURL := fmt.Sprintf("/sps/oauth/oauth20/authorize?client_id=wdp-eam-wbc-customer-pkce"+
		"&code_challenge=%s&code_challenge_method=S256"+
		"&nonce=%s&redirect_uri=%s"+
		"&response_type=code&scope=openid&state=%s",
		challenge, nonce,
		url.QueryEscape("https://www.ui.westpac.com.au/static/scripts/oidc/silent-renew.html"),
		state)

	req, err := http.NewRequest("GET", c.baseURL+authURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating authorize request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("step 4: %w", err)
	}
	defer resp.Body.Close()
	authBody, _ := io.ReadAll(resp.Body)

	location := resp.Header.Get("Location")
	c.log(fmt.Sprintf("[AUTH] Step 4: status=%d, location=%s", resp.StatusCode, truncate(location, 200)))

	var code string
	if location != "" {
		if u, err := url.Parse(location); err == nil {
			code = u.Query().Get("code")
		}
	}
	if code == "" {
		code = extractCodeFromHTML(string(authBody))
	}

	// Step 5: Token Exchange
	if code != "" {
		c.log(fmt.Sprintf("[AUTH] Step 5: token exchange (code=%s)", truncate(code, 30)))
		tokenData := url.Values{
			"client_id":     {"wdp-eam-wbc-customer-pkce"},
			"code":          {code},
			"code_verifier": {verifier},
			"grant_type":    {"authorization_code"},
			"redirect_uri":  {"https://www.ui.westpac.com.au/static/scripts/oidc/silent-renew.html"},
		}
		status, body, err = c.doForm("POST", "/spsu/sps/oauth/oauth20/token", tokenData.Encode(), nil)
		if err != nil {
			c.log(fmt.Sprintf("[AUTH] Step 5 error: %v", err))
		} else {
			c.log(fmt.Sprintf("[AUTH] Step 5: status=%d, body=%s", status, truncate(string(body), 500)))
		}
	} else {
		c.log("[AUTH] No auth code, skipping token exchange")
	}

	// Step 6: Select profile
	c.log("[AUTH] Step 6: selectProfile")
	profileBody := map[string]interface{}{
		"isSignInProfile": true,
		"profileType":     "Personal",
	}
	status, body, err = c.doJSON("POST", "/secure/wbcwebapi/api/profile/v1/selectProfile", profileBody, nil)
	if err != nil {
		c.log(fmt.Sprintf("[AUTH] Step 6 error: %v", err))
	} else {
		c.log(fmt.Sprintf("[AUTH] Step 6: status=%d, body=%s", status, truncate(string(body), 500)))
	}

	result := &AuthResult{}
	if len(body) > 0 {
		data := models.UnwrapData(body)
		var profileResp models.RawResponse
		models.ParseFlexible(data, &profileResp)
		if pid, ok := profileResp["currentProfileID"]; ok {
			result.ProfileID = fmt.Sprintf("%v", pid)
		}
		c.log(fmt.Sprintf("[AUTH] Profile data keys: %v", func() []string {
			keys := make([]string, 0, len(profileResp))
			for k := range profileResp {
				keys = append(keys, k)
			}
			return keys
		}()))
	}

	c.mu.RLock()
	result.XSRFToken = c.xsrfToken
	c.mu.RUnlock()

	c.log(fmt.Sprintf("[AUTH] Done. ProfileID=%s, XSRF=%v", result.ProfileID, result.XSRFToken != ""))
	return result, nil
}

func (c *Client) SendAppCapabilities() error {
	caps := map[string]string{
		"favouriteCustomisation": "1.0",
		"paymentshub":            "1.0",
		"transactionDetail":      "1.0",
		"personalisedinsights":   "4.0",
		"activatecard":           "2.0",
		"magicsearch":            "1.0",
		"woisso":                 "2.0",
	}
	status, _, err := c.doJSON("POST", "/secure/wbcwebapi/api/deviceManagement/v1/sendAppCapabilities", caps, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("sendAppCapabilities returned %d", status)
	}
	return nil
}

func (c *Client) log(msg string) {
	if c.debug && c.debugLog != nil {
		c.debugLog(msg)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func extractCodeFromHTML(body string) string {
	if idx := strings.Index(body, "code="); idx != -1 {
		end := strings.IndexAny(body[idx+5:], "&\"' >")
		if end == -1 {
			end = len(body[idx+5:])
		}
		if end > 0 {
			return body[idx+5 : idx+5+end]
		}
	}
	return ""
}
