package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

type SummaryModel struct {
	netWorth      *models.NetWorth
	spendAnalysis *models.SpendAnalysis
	rawNetWorth   []byte
	rawSpend      []byte
	loading       bool
	errorMsg      string
	spinner       spinner.Model
	width         int
	height        int
}

func NewSummaryModel() SummaryModel {
	return SummaryModel{
		spinner: NewSpinner(),
	}
}

func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	if m.loading {
		if smsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(smsg)
			return m, cmd
		}
	}
	return m, nil
}

func (m SummaryModel) View() string {
	var b strings.Builder

	b.WriteString(RenderHeader(" Financial Summary ", m.width))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(RenderLoading(m.spinner, "Loading financial data..."))
		return b.String()
	}

	if m.errorMsg != "" {
		b.WriteString(RenderError(m.errorMsg))
		b.WriteString("\n\n")
	}

	// Net Worth
	b.WriteString(SubtitleStyle.Render("Net Worth"))
	b.WriteString("\n\n")

	if m.netWorth != nil {
		b.WriteString(RenderKeyValue("Assets", RenderMoney(m.netWorth.TotalAssets)))
		b.WriteString("\n")
		b.WriteString(RenderKeyValue("Liabilities", RenderMoney(m.netWorth.TotalLiabilities)))
		b.WriteString("\n")
		b.WriteString(RenderKeyValue("Net Worth", RenderMoney(m.netWorth.TotalNetWorth)))
		b.WriteString("\n")
	} else if m.rawNetWorth != nil {
		b.WriteString(renderRaw(m.rawNetWorth))
	} else {
		b.WriteString(DimStyle.Render("No data"))
	}

	b.WriteString("\n\n")

	// Spend Analysis
	b.WriteString(SubtitleStyle.Render("Spending Analysis"))
	b.WriteString("\n\n")

	if m.spendAnalysis != nil {
		cf := m.spendAnalysis.CashflowSummary
		b.WriteString(RenderKeyValue("Period", cf.DateRangeLabel))
		b.WriteString("\n")
		b.WriteString(RenderKeyValue("Income", RenderMoney(cf.Income.Value)))
		b.WriteString("\n")
		b.WriteString(RenderKeyValue("Expenses", RenderMoney(cf.Expenses.Value)))
		b.WriteString("\n")
		b.WriteString(RenderKeyValue("Cash Flow", RenderMoney(cf.Amount)))
		b.WriteString("\n")

		// Expense categories
		cats := m.spendAnalysis.CategoriesSummary.Tabs.Expenses.Categories
		if len(cats) > 0 {
			b.WriteString("\n")
			b.WriteString(DimStyle.Render("Top Expenses:"))
			b.WriteString("\n")
			for _, cat := range cats {
				b.WriteString(fmt.Sprintf("  %-25s  %s\n",
					utils.TruncateString(cat.CategoryName, 25),
					RenderMoney(cat.Value)))
			}
		}

		// Income categories
		icats := m.spendAnalysis.CategoriesSummary.Tabs.Income.Categories
		if len(icats) > 0 {
			b.WriteString("\n")
			b.WriteString(DimStyle.Render("Income:"))
			b.WriteString("\n")
			for _, cat := range icats {
				b.WriteString(fmt.Sprintf("  %-25s  %s\n",
					utils.TruncateString(cat.CategoryName, 25),
					RenderMoney(cat.Value)))
			}
		}
	} else if m.rawSpend != nil {
		b.WriteString(renderRaw(m.rawSpend))
	} else {
		b.WriteString(DimStyle.Render("No data"))
	}

	b.WriteString("\n")

	return b.String()
}

func renderRaw(data []byte) string {
	var pretty interface{}
	if json.Unmarshal(data, &pretty) == nil {
		formatted, _ := json.MarshalIndent(pretty, "", "  ")
		raw := string(formatted)
		if len(raw) > 1000 {
			raw = raw[:1000] + "..."
		}
		return DimStyle.Render(raw)
	}
	return DimStyle.Render(string(data))
}

func (m SummaryModel) SetNetWorth(nw *models.NetWorth, raw []byte) SummaryModel {
	m.netWorth = nw
	m.rawNetWorth = raw
	return m
}

func (m SummaryModel) SetSpendAnalysis(sa *models.SpendAnalysis, raw []byte) SummaryModel {
	m.spendAnalysis = sa
	m.rawSpend = raw
	return m
}

func (m SummaryModel) SetLoading(loading bool) SummaryModel {
	m.loading = loading
	return m
}

func (m SummaryModel) SetError(err error) SummaryModel {
	if err != nil {
		m.errorMsg = fmt.Sprintf("%v", err)
	} else {
		m.errorMsg = ""
	}
	m.loading = false
	return m
}

func (m SummaryModel) SetSize(w, h int) SummaryModel {
	m.width = w
	m.height = h
	return m
}
