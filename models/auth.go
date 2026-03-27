package models

import "time"

type SessionData struct {
	DeviceID    string    `json:"device_id"`
	CustomerID  string    `json:"customer_id"`
	ProfileID   string    `json:"profile_id"`
	AccessToken string    `json:"access_token"`
	XSRFToken   string    `json:"xsrf_token"`
	Cookies     []Cookie  `json:"cookies"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

func (s *SessionData) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().After(s.ExpiresAt)
}

type PKCEParams struct {
	Verifier  string
	Challenge string
	State     string
	Nonce     string
}
