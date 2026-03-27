package api

import (
	"encoding/json"
	"fmt"

	"github.com/noma4i/westpac-cli/models"
)

func (c *Client) GetAccounts() ([]models.Account, []byte, error) {
	status, body, err := c.doGet("/secure/wbcwebapi/api/accounts/v5/userAccounts?canShowCards=true", nil)
	if err != nil {
		return nil, body, err
	}
	if status >= 400 {
		return nil, body, fmt.Errorf("getAccounts returned %d", status)
	}

	data := models.UnwrapData(body)

	var resp models.AccountsDataResponse
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.AllAccounts) > 0 {
		c.log(fmt.Sprintf("[ACCOUNTS] parsed %d accounts", len(resp.AllAccounts)))
		return resp.AllAccounts, body, nil
	}

	c.log(fmt.Sprintf("[ACCOUNTS] could not parse allAccounts, data: %s", truncate(string(data), 300)))
	return nil, body, fmt.Errorf("could not parse accounts response")
}

func (c *Client) GetAccountDetails(accountID string) (*models.AccountDetail, []byte, error) {
	path := fmt.Sprintf("/secure/wbcwebapi/api/accounts/%s/v7/details", accountID)
	status, body, err := c.doGetWithHeaders(path, map[string]string{
		"x-profileExperience": "Native",
	}, nil)
	if err != nil {
		return nil, body, err
	}
	if status >= 400 {
		return nil, body, fmt.Errorf("getAccountDetails returned %d", status)
	}

	data := models.UnwrapData(body)
	var detail models.AccountDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return nil, body, fmt.Errorf("parsing account detail: %w", err)
	}

	return &detail, body, nil
}
