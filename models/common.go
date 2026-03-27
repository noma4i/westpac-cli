package models

import (
	"encoding/json"
	"fmt"
)

type RawResponse map[string]interface{}

// UnwrapData extracts the "data" field from Westpac API responses.
// All API responses are wrapped: {"data": {...actual data...}}
func UnwrapData(body []byte) []byte {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Data != nil {
		return []byte(envelope.Data)
	}
	return body
}

type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (m Money) String() string {
	if m.Currency == "" {
		m.Currency = "AUD"
	}
	sign := ""
	amt := m.Amount
	if amt < 0 {
		sign = "-"
		amt = -amt
	}
	return fmt.Sprintf("%s$%.2f", sign, amt)
}

func ParseFlexible(data []byte, target interface{}) (RawResponse, error) {
	if err := json.Unmarshal(data, target); err != nil {
		var raw RawResponse
		if rawErr := json.Unmarshal(data, &raw); rawErr != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return raw, nil
	}
	var raw RawResponse
	json.Unmarshal(data, &raw)
	return raw, nil
}
