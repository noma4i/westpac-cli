package models

// Actual API response (inside "data"):
// {"totalNetWorth":-7778.8,"totalAssets":2549.06,"totalLiabilities":-10327.86,"barGraphPercentage":19.8}
type NetWorth struct {
	TotalNetWorth    float64     `json:"totalNetWorth"`
	TotalAssets      float64     `json:"totalAssets"`
	TotalLiabilities float64     `json:"totalLiabilities"`
	BarGraphPct      float64     `json:"barGraphPercentage"`
	Extra            RawResponse `json:"-"`
}

// Actual API response (inside "data"):
// {"cashflowSummary":{"title":"Cash flow","amount":-7877.26,
//   "income":{"value":35262.96,"body":"Income"},
//   "expenses":{"value":-43140.22,"body":"Expenses"}},
//  "categoriesSummary":{"tabs":{"income":{"categories":[...]},
//   "expenses":{"categories":[...]}}}}
type SpendAnalysis struct {
	CashflowSummary   CashflowSummary   `json:"cashflowSummary"`
	CategoriesSummary CategoriesSummary  `json:"categoriesSummary"`
	Extra             RawResponse        `json:"-"`
}

type CashflowSummary struct {
	Title          string          `json:"title"`
	DateRangeLabel string          `json:"dateRangeLabel"`
	Amount         float64         `json:"amount"`
	Income         CashflowEntry   `json:"income"`
	Expenses       CashflowEntry   `json:"expenses"`
}

type CashflowEntry struct {
	Value float64 `json:"value"`
	Body  string  `json:"body"`
}

type CategoriesSummary struct {
	Tabs           CategoriesTabs `json:"tabs"`
	DateRangeLabel string         `json:"dateRangeLabel"`
}

type CategoriesTabs struct {
	Income   CategoryTab `json:"income"`
	Expenses CategoryTab `json:"expenses"`
}

type CategoryTab struct {
	FilterLabel string          `json:"filterLabel"`
	Categories  []SpendCategory `json:"categories"`
}

type SpendCategory struct {
	CategoryID   int     `json:"categoryID"`
	CategoryName string  `json:"categoryName"`
	Icon         string  `json:"icon"`
	Value        float64 `json:"value"`
}
