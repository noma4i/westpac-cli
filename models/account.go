package models

type Account struct {
	AccountID        string  `json:"accountID"`
	DisplayName      string  `json:"displayName"`
	AccountNumber    string  `json:"accountNumber"`
	BSB              string  `json:"bsb,omitempty"`
	AvailableBalance float64 `json:"availableBalance"`
	CurrentBalance   float64 `json:"currentBalance"`
	IsFavourite      bool    `json:"isFavourite"`
	IsHidden         bool    `json:"isHidden"`
	IsNew            bool    `json:"isNew"`
}

type AccountsDataResponse struct {
	AllAccounts []Account `json:"allAccounts"`
}

type AccountDetail struct {
	Account
	InterestRate float64 `json:"interestRate,omitempty"`
	OpenDate     string  `json:"openDate,omitempty"`
}
