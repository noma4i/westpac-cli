package api

import (
	"encoding/json"
	"fmt"

	"github.com/noma4i/westpac-cli/models"
)

func (c *Client) GetNetWorth() (*models.NetWorth, []byte, error) {
	status, body, err := c.doGet("/secure/wbcwebapi/api/financialManagement/networth/v1/summary", nil)
	if err != nil {
		return nil, body, err
	}
	if status >= 400 {
		return nil, body, fmt.Errorf("getNetWorth returned %d", status)
	}

	data := models.UnwrapData(body)
	var nw models.NetWorth
	if err := json.Unmarshal(data, &nw); err != nil {
		c.log(fmt.Sprintf("[PARSE_ERR] networth: %v", err))
		return nil, body, nil
	}
	return &nw, body, nil
}

func (c *Client) GetSpendAnalysis() (*models.SpendAnalysis, []byte, error) {
	status, body, err := c.doGet("/secure/wbcwebapi/api/financialmanagement/spendAnalysis/v2/summary", nil)
	if err != nil {
		return nil, body, err
	}
	if status >= 400 {
		return nil, body, fmt.Errorf("getSpendAnalysis returned %d", status)
	}

	data := models.UnwrapData(body)
	var sa models.SpendAnalysis
	if err := json.Unmarshal(data, &sa); err != nil {
		c.log(fmt.Sprintf("[PARSE_ERR] spendAnalysis: %v", err))
		return nil, body, nil
	}
	return &sa, body, nil
}

func (c *Client) GetAlerts() ([]byte, error) {
	status, body, err := c.doGet("/secure/wbcwebapi/api/messaging/v1/skinnyalerts", nil)
	if err != nil {
		return body, err
	}
	if status >= 400 {
		return body, fmt.Errorf("getAlerts returned %d", status)
	}
	return body, nil
}
