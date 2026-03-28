package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

const PageSize = 75

type TransactionsModel struct {
	account      models.Account
	transactions []models.Transaction
	filtered     []models.Transaction
	cursor       int
	scrollOffset int
	page         int
	hasMore      bool
	loading      bool
	loadingMore  bool
	errorMsg     string
	spinner      spinner.Model
	width        int
	height       int
	search       string
	searching    bool
}

func NewTransactionsModel() TransactionsModel {
	return TransactionsModel{
		spinner: NewSpinner(),
	}
}

type TxnSelectedMsg struct {
	Transaction models.Transaction
	AccountID   string
}

type TxnLoadPageMsg struct {
	AccountID string
	Page      int
}

type TxnPageLoadedMsg struct {
	Transactions []models.Transaction
}

func (m TransactionsModel) visibleList() []models.Transaction {
	if m.searching && m.search != "" {
		return m.filtered
	}
	return m.transactions
}

func (m *TransactionsModel) applyFilter() {
	if m.search == "" {
		m.filtered = nil
		return
	}
	q := strings.ToLower(m.search)
	m.filtered = nil
	for _, tx := range m.transactions {
		if strings.Contains(strings.ToLower(tx.Description), q) ||
			strings.Contains(strings.ToLower(tx.Category), q) ||
			strings.Contains(utils.FormatMoney(tx.Amount, "AUD"), q) {
			m.filtered = append(m.filtered, tx)
		}
	}
}

func (m TransactionsModel) viewHeight() int {
	// header(1) + marginBottom(1) + "\n"(1) + account info(1) + "\n\n"(2) + col header(1) + separator(1) + "\n"(1) + footer(1) + "\n"(1) = 11
	h := m.height - 11
	if m.searching {
		h -= 2
	}
	if h < 5 {
		h = 5
	}
	return h
}

func (m TransactionsModel) Update(msg tea.Msg) (TransactionsModel, tea.Cmd) {
	if m.loading {
		if smsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(smsg)
			return m, cmd
		}
		return m, nil
	}

	if m.loadingMore {
		if smsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(smsg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if m.searching {
			switch key {
			case "esc":
				m.searching = false
				m.search = ""
				m.filtered = nil
				m.cursor = 0
				m.scrollOffset = 0
				return m, nil
			case "enter":
				list := m.visibleList()
				if len(list) > 0 && m.cursor < len(list) {
					tx := list[m.cursor]
					return m, func() tea.Msg {
						return TxnSelectedMsg{Transaction: tx, AccountID: m.account.AccountID}
					}
				}
				return m, nil
			case "up":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.scrollOffset {
						m.scrollOffset = m.cursor
					}
				}
				return m, nil
			case "down":
				list := m.visibleList()
				if m.cursor < len(list)-1 {
					m.cursor++
					vh := m.viewHeight()
					if m.cursor >= m.scrollOffset+vh {
						m.scrollOffset = m.cursor - vh + 1
					}
				}
				return m, nil
			case "backspace":
				if len(m.search) > 0 {
					m.search = m.search[:len(m.search)-1]
					m.applyFilter()
					m.cursor = 0
					m.scrollOffset = 0
				}
				if m.search == "" {
					m.searching = false
					m.filtered = nil
				}
				return m, nil
			default:
				if len(key) == 1 && key >= " " {
					m.search += key
					m.applyFilter()
					m.cursor = 0
					m.scrollOffset = 0
					return m, nil
				}
			}
			return m, nil
		}

		// Normal mode
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down", "j":
			list := m.visibleList()
			if m.cursor < len(list)-1 {
				m.cursor++
				vh := m.viewHeight()
				if m.cursor >= m.scrollOffset+vh {
					m.scrollOffset = m.cursor - vh + 1
				}
				// Auto-load more when near the end
				if m.hasMore && !m.loadingMore && m.cursor >= len(m.transactions)-10 {
					m.page++
					m.loadingMore = true
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							return TxnLoadPageMsg{AccountID: m.account.AccountID, Page: m.page}
						},
					)
				}
			}
		case "enter":
			list := m.visibleList()
			if len(list) > 0 && m.cursor < len(list) {
				tx := list[m.cursor]
				return m, func() tea.Msg {
					return TxnSelectedMsg{Transaction: tx, AccountID: m.account.AccountID}
				}
			}
		default:
			// Start searching on any printable character
			if len(key) == 1 && key >= " " {
				m.searching = true
				m.search = key
				m.applyFilter()
				m.cursor = 0
				m.scrollOffset = 0
				return m, nil
			}
		}
	}
	return m, nil
}

func (m TransactionsModel) View() string {
	var b strings.Builder

	accName := m.account.DisplayName
	if accName == "" {
		accName = "Account"
	}

	title := fmt.Sprintf(" %s ", accName)
	b.WriteString(RenderHeader(title, m.width))
	b.WriteString("\n")

	// Account info
	balStr := utils.FormatMoney(m.account.CurrentBalance, "AUD")
	availStr := utils.FormatMoney(m.account.AvailableBalance, "AUD")
	acctLine := MoneyStyle(m.account.CurrentBalance).Render(balStr)
	if m.account.AvailableBalance != m.account.CurrentBalance {
		acctLine += DimStyle.Render("  available: ") + ValueStyle.Render(availStr)
	}
	if m.account.BSB != "" {
		bsb := m.account.BSB
		accNo := m.account.AccountNumber
		if utils.MaskMode {
			bsb = utils.MaskStars(bsb)
			accNo = utils.MaskStars(accNo)
		}
		acctLine += DimStyle.Render(fmt.Sprintf("  BSB %s  Acc %s", bsb, accNo))
	}
	b.WriteString(" " + acctLine)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(RenderLoading(m.spinner, "Loading transactions..."))
		return b.String()
	}

	if m.errorMsg != "" {
		b.WriteString(RenderError(m.errorMsg))
		b.WriteString("\n\n")
		return b.String()
	}

	// Search bar
	if m.searching {
		searchLine := fmt.Sprintf(" Search: %s_", m.search)
		b.WriteString(SubtitleStyle.Render(searchLine))
		b.WriteString("\n\n")
	}

	list := m.visibleList()

	if len(list) == 0 {
		if m.searching {
			b.WriteString(DimStyle.Render("No matches"))
		} else {
			b.WriteString(DimStyle.Render("No transactions found"))
		}
		b.WriteString("\n")
		return b.String()
	}

	// Column widths
	// format: " %-dateW  %-descW  %amtW" inside NormalStyle(Padding 0,1)
	// total = 1 + dateW + 2 + descW + 2 + amtW + 2(padding) = descW + 33
	dateW := 12
	amtW := 14
	descW := m.width - dateW - amtW - 9 // 7 + 2 right margin
	if descW < 20 {
		descW = 20
	}

	// Header row
	headerLine := fmt.Sprintf(" %-*s  %-*s  %*s", dateW, "Date", descW, "Description", amtW, "Amount")
	b.WriteString(DimStyle.Render(headerLine))
	b.WriteString("\n")
	b.WriteString(DimStyle.Render(strings.Repeat("-", m.width-2)))
	b.WriteString("\n")

	vh := m.viewHeight()
	end := m.scrollOffset + vh
	if end > len(list) {
		end = len(list)
	}
	visible := list[m.scrollOffset:end]

	for i, tx := range visible {
		idx := m.scrollOffset + i
		date := utils.FormatUnixDate(tx.Date)
		desc := utils.TruncateString(utils.MaskPartial(tx.Description, 5), descW)
		amtStr := utils.FormatMoney(tx.Amount, "AUD")

		if idx == m.cursor {
			line := fmt.Sprintf("%-*s  %-*s  %*s", dateW, date, descW, desc, amtW, amtStr)
			b.WriteString(SelectedStyle.Width(m.width).Render(line))
		} else {
			datePart := fmt.Sprintf("%-*s  %-*s  ", dateW, date, descW, desc)
			amtPart := MoneyStyle(tx.Amount).Render(fmt.Sprintf("%*s", amtW, amtStr))
			b.WriteString(NormalStyle.Render(datePart + amtPart))
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	info := fmt.Sprintf("%d transactions", len(m.transactions))
	if m.loadingMore {
		info += fmt.Sprintf("  %s loading more...", m.spinner.View())
	} else if m.hasMore {
		info += " (scroll for more)"
	}
	b.WriteString(DimStyle.Render(info))
	b.WriteString("\n")

	return b.String()
}

func (m TransactionsModel) SetAccount(acc models.Account) TransactionsModel {
	m.account = acc
	m.page = 0
	m.cursor = 0
	m.scrollOffset = 0
	m.transactions = nil
	m.filtered = nil
	m.search = ""
	m.searching = false
	return m
}

func (m TransactionsModel) SetTransactions(txns []models.Transaction) TransactionsModel {
	m.transactions = txns
	m.loading = false
	m.loadingMore = false
	m.errorMsg = ""
	m.hasMore = len(txns) >= PageSize
	m.cursor = 0
	m.scrollOffset = 0
	return m
}

func (m TransactionsModel) AppendTransactions(txns []models.Transaction) TransactionsModel {
	m.transactions = append(m.transactions, txns...)
	m.loadingMore = false
	m.hasMore = len(txns) >= PageSize
	return m
}

func (m TransactionsModel) SetLoading(loading bool) TransactionsModel {
	m.loading = loading
	return m
}

func (m TransactionsModel) SetError(err error) TransactionsModel {
	if err != nil {
		m.errorMsg = fmt.Sprintf("%v", err)
	} else {
		m.errorMsg = ""
	}
	m.loading = false
	m.loadingMore = false
	return m
}

func (m TransactionsModel) SetSize(w, h int) TransactionsModel {
	m.width = w
	m.height = h
	return m
}

func (m TransactionsModel) IsSearching() bool {
	return m.searching
}

func (m TransactionsModel) ClearSearch() TransactionsModel {
	m.searching = false
	m.search = ""
	m.filtered = nil
	m.cursor = 0
	m.scrollOffset = 0
	return m
}

func (m TransactionsModel) AccountID() string {
	return m.account.AccountID
}

func (m TransactionsModel) Page() int {
	return m.page
}
