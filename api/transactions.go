package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/noma4i/westpac-cli/models"
)

func (c *Client) GetTransactions(accountID string, page, pageSize int) ([]models.Transaction, []byte, error) {
	from := page * pageSize
	to := from + pageSize - 1

	body := map[string]interface{}{
		"context":  "accountSpecific",
		"accounts": []string{accountID},
		"term":     "",
		"type":     "",
	}

	headers := map[string]string{
		"Range":               fmt.Sprintf("items=%d-%d", from, to),
		"x-profileExperience": "Native",
	}

	status, respBody, err := c.doPostWithHeaders(
		"/secure/wbcwebapi/api/transactions/v3/transactions",
		headers, body, nil,
	)
	if err != nil {
		return nil, respBody, err
	}
	if status >= 400 {
		return nil, respBody, fmt.Errorf("getTransactions returned %d", status)
	}

	data := models.UnwrapData(respBody)

	var resp models.TransactionsResponse
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.Transactions) > 0 {
		c.log(fmt.Sprintf("[TRANSACTIONS] parsed %d transactions", len(resp.Transactions)))
		return resp.Transactions, respBody, nil
	}

	var txns []models.Transaction
	if err := json.Unmarshal(data, &txns); err == nil && len(txns) > 0 {
		return txns, respBody, nil
	}

	return nil, respBody, fmt.Errorf("could not parse transactions response")
}

// GetTransactionDetail fetches detail using params from the transaction's detailsLink.
func (c *Client) GetTransactionDetail(tx models.Transaction) (*models.TransactionDetail, []byte, error) {
	// Parse params from detailsLink (the source of truth for all query params)
	params := parseDetailsLink(tx.DetailsLink)

	txID := params.Get("transactionId")
	if txID == "" {
		txID = tx.TransactionID
	}

	path := fmt.Sprintf("/secure/wbcwebapi/api/transactions/v2/%s/details?accountId=%s&accountType=%s&context=%s&isIntraday=%s&transactionDate=%s",
		txID,
		params.Get("accountId"),
		params.Get("accountType"),
		params.Get("context"),
		params.Get("isIntraday"),
		params.Get("transactionDate"),
	)

	c.log(fmt.Sprintf("[TX_DETAIL] %s", path))

	status, respBody, err := c.doGet(path, nil)
	if err != nil {
		return nil, respBody, err
	}
	if status >= 400 {
		return nil, respBody, fmt.Errorf("transaction detail returned %d", status)
	}

	data := models.UnwrapData(respBody)
	detail := models.ParseTransactionDetail(data)
	if detail == nil {
		return nil, respBody, fmt.Errorf("could not parse transaction detail")
	}

	return detail, respBody, nil
}

func parseDetailsLink(link string) url.Values {
	if link == "" {
		return url.Values{}
	}
	u, err := url.Parse(link)
	if err != nil {
		return url.Values{}
	}
	return u.Query()
}

func (c *Client) GetSimilarCount(txID string) (int, error) {
	path := fmt.Sprintf("/secure/wbcwebapi/api/transactions/%s/v1/similarCount", txID)
	var result struct {
		Count int `json:"count"`
	}
	status, _, err := c.doGet(path, &result)
	if err != nil {
		return 0, err
	}
	if status >= 400 {
		return 0, fmt.Errorf("getSimilarCount returned %d", status)
	}
	return result.Count, nil
}
