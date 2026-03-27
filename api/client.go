package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/noma4i/westpac-cli/utils"
)

const (
	BaseURL    = "https://banking.westpac.com.au"
	UserAgent  = "IONBanking-WBC/12.0.0(iPhone OS; 26.4; APPLE; iPhone18,2; en_AU)"
	SystemID   = "A00931"
	BrandSilo  = "WPAC"
	ChannelTyp = "Mobile"
)

type Client struct {
	http      *http.Client
	baseURL   string
	deviceID  string
	xsrfToken string
	mu        sync.RWMutex
	debug     bool
	debugLog  func(string)
}

type westpacTransport struct {
	base   http.RoundTripper
	client *Client
}

func (t *westpacTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-channelType", ChannelTyp)
	req.Header.Set("brandSilo", BrandSilo)
	req.Header.Set("x-originatingDeviceId", t.client.deviceID)
	req.Header.Set("x-messageId", utils.NewUUID())
	req.Header.Set("x-appCorrelationId", utils.NewUUID())
	req.Header.Set("x-originatingSystemId", SystemID)
	req.Header.Set("x-consumerType", "Customer")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-AU;q=1.0")
	req.Header.Set("ADRUM", "isAjax:true")
	req.Header.Set("ADRUM_1", "isMobile:true")

	if strings.Contains(req.URL.Path, "/secure/") {
		req.Header.Set("x-profileExperience", "Native")
	}

	t.client.mu.RLock()
	xsrf := t.client.xsrfToken
	t.client.mu.RUnlock()

	if req.Method == "POST" && xsrf != "" {
		req.Header.Set("x-XSRF", xsrf)
	}

	if t.client.debug && t.client.debugLog != nil {
		t.client.debugLog(fmt.Sprintf("[REQ] %s %s", req.Method, req.URL.String()))
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if xsrfHeader := resp.Header.Get("x-XSRF"); xsrfHeader != "" {
		t.client.mu.Lock()
		t.client.xsrfToken = xsrfHeader
		t.client.mu.Unlock()
	}

	return resp, nil
}

func NewClient(deviceID string, debug bool, debugLog func(string)) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	if deviceID == "" {
		deviceID = strings.ReplaceAll(strings.ToUpper(utils.NewUUID()), "-", "")
	}

	c := &Client{
		baseURL:  BaseURL,
		deviceID: deviceID,
		debug:    debug,
		debugLog: debugLog,
	}

	c.http = &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &westpacTransport{
			base:   http.DefaultTransport,
			client: c,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return c, nil
}

func (c *Client) SetXSRFToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.xsrfToken = token
}

func (c *Client) DeviceID() string {
	return c.deviceID
}

func (c *Client) do(req *http.Request) (int, []byte, error) {
	path := req.URL.Path

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug && c.debugLog != nil {
		c.debugLog(fmt.Sprintf("[RESP] %d %s (%d bytes)", resp.StatusCode, path, len(body)))
		if len(body) > 0 {
			logBody := string(body)
			if len(logBody) > 10000 {
				logBody = logBody[:10000] + "...[truncated]"
			}
			c.debugLog(fmt.Sprintf("[BODY] %s", logBody))
		}
	}

	if err := c.checkHTMLRedirect(body, path); err != nil {
		return resp.StatusCode, body, err
	}

	return resp.StatusCode, body, nil
}

func (c *Client) doJSON(method, path string, reqBody interface{}, result interface{}) (int, []byte, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return 0, nil, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return 0, nil, fmt.Errorf("creating request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	status, body, err := c.do(req)
	if err != nil {
		return status, body, err
	}

	if result != nil && len(body) > 0 {
		if uerr := json.Unmarshal(body, result); uerr != nil {
			if c.debug && c.debugLog != nil {
				c.debugLog(fmt.Sprintf("[PARSE_ERR] %s: %v", path, uerr))
			}
		}
	}

	return status, body, nil
}

func (c *Client) doForm(method, path string, formData string, result interface{}) (int, []byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, strings.NewReader(formData))
	if err != nil {
		return 0, nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	status, body, err := c.do(req)
	if err != nil {
		return status, body, err
	}

	if result != nil && len(body) > 0 {
		if uerr := json.Unmarshal(body, result); uerr != nil {
			if c.debug && c.debugLog != nil {
				c.debugLog(fmt.Sprintf("[PARSE_ERR] %s: %v", path, uerr))
			}
		}
	}

	return status, body, nil
}

func (c *Client) doGet(path string, result interface{}) (int, []byte, error) {
	return c.doJSON("GET", path, nil, result)
}

func (c *Client) doGetWithHeaders(path string, headers map[string]string, result interface{}) (int, []byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("creating request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	status, body, err := c.do(req)
	if err != nil {
		return status, body, err
	}

	if result != nil && len(body) > 0 {
		json.Unmarshal(body, result)
	}

	return status, body, nil
}

func (c *Client) doPostWithHeaders(path string, headers map[string]string, reqBody interface{}, result interface{}) (int, []byte, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return 0, nil, err
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	status, body, err := c.do(req)
	if err != nil {
		return status, body, err
	}

	if result != nil && len(body) > 0 {
		json.Unmarshal(body, result)
	}

	return status, body, nil
}

// followRedirects follows a chain of 302 redirects collecting cookies along the way.
// Stops after 10 redirects or when getting a non-redirect response.
func (c *Client) followRedirects(location string) error {
	for i := 0; i < 10; i++ {
		// Resolve relative URLs
		if !strings.HasPrefix(location, "http") {
			location = c.baseURL + location
		}

		req, err := http.NewRequest("GET", location, nil)
		if err != nil {
			return fmt.Errorf("redirect request: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("redirect: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if c.debug && c.debugLog != nil {
			c.debugLog(fmt.Sprintf("[REDIRECT] %d %s -> %s (%d bytes)",
				resp.StatusCode, location, resp.Header.Get("Location"), len(body)))
		}

		if resp.StatusCode != 301 && resp.StatusCode != 302 && resp.StatusCode != 303 {
			return nil
		}

		location = resp.Header.Get("Location")
		if location == "" {
			return nil
		}
	}
	return fmt.Errorf("too many redirects")
}

func (c *Client) checkHTMLRedirect(body []byte, path string) error {
	if len(body) == 0 {
		return nil
	}
	s := strings.TrimSpace(string(body))
	if strings.HasPrefix(s, "<!DOCTYPE") || strings.HasPrefix(s, "<html") || strings.HasPrefix(s, "<HTML") {
		if c.debug && c.debugLog != nil {
			c.debugLog(fmt.Sprintf("[HTML_REDIRECT] %s: session expired or auth required", path))
		}
		return fmt.Errorf("session expired (got HTML redirect)")
	}
	return nil
}
