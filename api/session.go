package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/noma4i/westpac-cli/models"
)

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "westpac"), nil
}

func sessionPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func (c *Client) SaveSession(customerID, profileID string) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// Collect cookies from the jar
	baseURL, _ := url.Parse(c.baseURL)
	httpCookies := c.http.Jar.Cookies(baseURL)

	cookies := make([]models.Cookie, len(httpCookies))
	for i, hc := range httpCookies {
		cookies[i] = models.Cookie{
			Name:   hc.Name,
			Value:  hc.Value,
			Domain: hc.Domain,
			Path:   hc.Path,
		}
	}

	c.mu.RLock()
	xsrf := c.xsrfToken
	c.mu.RUnlock()

	session := &models.SessionData{
		DeviceID:   c.deviceID,
		CustomerID: customerID,
		ProfileID:  profileID,
		XSRFToken:  xsrf,
		Cookies:    cookies,
		ExpiresAt:  time.Now().Add(30 * time.Minute),
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Client) LoadSession() (*models.SessionData, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session models.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if session.IsExpired() {
		return nil, fmt.Errorf("session expired")
	}

	// Restore cookies to the jar
	baseURL, _ := url.Parse(c.baseURL)
	httpCookies := make([]*http.Cookie, len(session.Cookies))
	for i, sc := range session.Cookies {
		httpCookies[i] = &http.Cookie{
			Name:   sc.Name,
			Value:  sc.Value,
			Domain: sc.Domain,
			Path:   sc.Path,
		}
	}
	c.http.Jar.SetCookies(baseURL, httpCookies)

	c.deviceID = session.DeviceID

	c.mu.Lock()
	c.xsrfToken = session.XSRFToken
	c.mu.Unlock()

	return &session, nil
}

func (c *Client) ValidateSession() bool {
	status, _, err := c.doGet("/secure/wbcwebapi/api/messaging/v1/skinnyalerts", nil)
	if err != nil || status >= 400 {
		return false
	}
	return true
}

func ClearSession() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
