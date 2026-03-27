package models

import "encoding/json"

// Real API format (inside "data.transactions[]"):
// {"transactionID":"uuid","date":1774789200,"accountID":"uuid",
//  "isPending":false,"description":"COLES...","category":"...",
//  "amount":-86.10,"balance":2453.41,"detailsLink":"https://..."}
type Transaction struct {
	TransactionID string  `json:"transactionID"`
	AccountID     string  `json:"accountID"`
	Description   string  `json:"description"`
	Category      string  `json:"category,omitempty"`
	Amount        float64 `json:"amount"`
	Balance       float64 `json:"balance"`
	Date          int64   `json:"date"`
	IsPending     bool    `json:"isPending"`
	DetailsLink   string  `json:"detailsLink,omitempty"`
}

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// TransactionDetail - parsed from the screen components API response.
// API returns UI components in screen.modulesList with automationId keys.
type TransactionDetail struct {
	MerchantName    string  `json:"merchantName"`
	Amount          float64 `json:"transactionAmount"`
	Narrative       string  `json:"transactionNarrative"`
	Date            string  // from TransactionDate component
	Address         string  // from MerchantAddressValue
	Website         string  // from MerchantWebsite body.text
	Phone           string  // from MerchantPhoneNo body.text
	Description     string  // from MerchantTransactionDescriptionValue
	AuthDate        string  // from AuthorisationDateValue
	AuthStatus      string  // from AuthorisationstatusValue
	TransactionID   string  // from TransactionIdValue
	LogoURL         string  // from MerchantIcon downloadURL
	Lat             float64 // from MapSnapshot
	Long            float64 // from MapSnapshot
	HasMap          bool
}

// ParseTransactionDetail extracts fields from the raw API response.
// Top-level fields: merchantName, transactionAmount, transactionNarrative.
// Component fields: extracted by automationId from screen.modulesList.
func ParseTransactionDetail(rawData []byte) *TransactionDetail {
	var top struct {
		MerchantName string  `json:"merchantName"`
		Amount       float64 `json:"transactionAmount"`
		Narrative    string  `json:"transactionNarrative"`
		Screen       struct {
			ModulesList []struct {
				Components []json.RawMessage `json:"components"`
			} `json:"modulesList"`
		} `json:"screen"`
	}
	if err := json.Unmarshal(rawData, &top); err != nil {
		return nil
	}

	d := &TransactionDetail{
		MerchantName: top.MerchantName,
		Amount:       top.Amount,
		Narrative:    top.Narrative,
	}

	for _, mod := range top.Screen.ModulesList {
		for _, raw := range mod.Components {
			parseComponent(raw, d)
		}
	}

	return d
}

func parseComponent(raw json.RawMessage, d *TransactionDetail) {
	var comp struct {
		ComponentType string `json:"componentType"`
		Model         struct {
			Text         string `json:"text"`
			DownloadURL  string `json:"downloadURL"`
			Lat          float64 `json:"lat"`
			Long         float64 `json:"long"`
			Body         *struct {
				Text string `json:"text"`
			} `json:"body"`
			Metadata struct {
				AutomationID string `json:"automationId"`
			} `json:"metadata"`
		} `json:"model"`
	}
	if err := json.Unmarshal(raw, &comp); err != nil {
		return
	}

	m := comp.Model
	switch m.Metadata.AutomationID {
	case "TransactionDate":
		d.Date = m.Text
	case "MerchantAddressValue":
		d.Address = m.Text
	case "MerchantWebsite":
		if m.Body != nil {
			d.Website = m.Body.Text
		}
	case "MerchantPhoneNo":
		if m.Body != nil {
			d.Phone = m.Body.Text
		}
	case "MerchantTransactionDescriptionValue":
		d.Description = m.Text
	case "AuthorisationDateValue":
		d.AuthDate = m.Text
	case "AuthorisationstatusValue":
		d.AuthStatus = m.Text
	case "TransactionIdValue":
		d.TransactionID = m.Text
	case "MerchantIcon":
		d.LogoURL = m.DownloadURL
	case "MapSnapshot":
		d.Lat = m.Lat
		d.Long = m.Long
		d.HasMap = true
	}
}
