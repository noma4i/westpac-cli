package views

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

type TxnDetailModel struct {
	transaction models.Transaction
	detail      *models.TransactionDetail
	rawBody     []byte
	accountID   string
	loading     bool
	errorMsg    string
	spinner     spinner.Model
	width       int
	height      int
	scrollY     int
	lines       []string
	logoASCII   string
	copied      bool
}

func NewTxnDetailModel() TxnDetailModel {
	return TxnDetailModel{
		spinner: NewSpinner(),
	}
}

func (m TxnDetailModel) Update(msg tea.Msg) (TxnDetailModel, tea.Cmd) {
	if m.loading {
		if smsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(smsg)
			return m, cmd
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
			maxScroll := m.calcMaxScroll()
			if m.scrollY < maxScroll {
				m.scrollY++
			}
		case "C":
			return m, m.copyToClipboard()
		}
	}
	return m, nil
}

type TxnCopiedMsg struct{}
type TxnCopyResetMsg struct{}

func (m TxnDetailModel) copyToClipboard() tea.Cmd {
	return func() tea.Msg {
		if m.detail == nil {
			return TxnCopiedMsg{}
		}
		d := m.detail
		var sb strings.Builder
		if d.MerchantName != "" {
			sb.WriteString(d.MerchantName + "\n")
		}
		sb.WriteString(fmt.Sprintf("Amount: %s\n", utils.FormatMoney(d.Amount, "AUD")))
		if d.Date != "" {
			sb.WriteString(fmt.Sprintf("Date: %s\n", d.Date))
		}
		if d.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", d.Description))
		} else if d.Narrative != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", d.Narrative))
		}
		if d.AuthStatus != "" {
			sb.WriteString(fmt.Sprintf("Status: %s\n", d.AuthStatus))
		}
		if d.Address != "" {
			sb.WriteString(fmt.Sprintf("Address: %s\n", d.Address))
		}
		if d.Website != "" {
			sb.WriteString(fmt.Sprintf("Website: %s\n", d.Website))
		}
		if d.Phone != "" {
			sb.WriteString(fmt.Sprintf("Phone: %s\n", d.Phone))
		}
		if d.TransactionID != "" {
			sb.WriteString(fmt.Sprintf("Transaction ID: %s\n", d.TransactionID))
		}
		clipboard.WriteAll(sb.String())
		return TxnCopiedMsg{}
	}
}

func (m TxnDetailModel) calcMaxScroll() int {
	viewHeight := m.height - 6
	if viewHeight < 5 {
		viewHeight = 5
	}
	if len(m.lines) > viewHeight {
		return len(m.lines) - viewHeight
	}
	return 0
}

func (m TxnDetailModel) buildLines() []string {
	var lines []string

	if m.detail != nil {
		d := m.detail

		if m.logoASCII != "" {
			for _, l := range strings.Split(m.logoASCII, "\n") {
				if l != "" {
					lines = append(lines, l)
				}
			}
			lines = append(lines, "")
		}

		if d.MerchantName != "" {
			lines = append(lines, SubtitleStyle.Render(d.MerchantName))
			lines = append(lines, "")
		}

		lines = append(lines, RenderKeyValue("Amount", utils.FormatMoney(d.Amount, "AUD")))

		if d.Date != "" {
			dateParts := strings.Split(d.Date, "\n")
			lines = append(lines, RenderKeyValue("Date", dateParts[0]))
			for _, dp := range dateParts[1:] {
				lines = append(lines, RenderKeyValue("", strings.TrimSpace(dp)))
			}
		}

		if d.Description != "" {
			lines = append(lines, RenderKeyValue("Description", utils.MaskPartial(d.Description, 5)))
		} else if d.Narrative != "" {
			lines = append(lines, RenderKeyValue("Description", utils.MaskPartial(d.Narrative, 5)))
		}

		if d.AuthStatus != "" {
			lines = append(lines, RenderKeyValue("Status", d.AuthStatus))
		}

		if d.AuthDate != "" {
			lines = append(lines, RenderKeyValue("Auth Date", d.AuthDate))
		}

		if d.Address != "" {
			lines = append(lines, "")
			lines = append(lines, DimStyle.Render("Merchant Info:"))
			lines = append(lines, RenderKeyValue("Address", d.Address))
		}

		if d.Website != "" {
			lines = append(lines, RenderKeyValue("Website", d.Website))
		}

		if d.Phone != "" {
			lines = append(lines, RenderKeyValue("Phone", d.Phone))
		}

		if d.HasMap {
			lines = append(lines, RenderKeyValue("Location", fmt.Sprintf("%.6f, %.6f", d.Lat, d.Long)))
		}

		if d.TransactionID != "" {
			lines = append(lines, "")
			lines = append(lines, RenderKeyValue("Transaction ID", utils.MaskText(d.TransactionID)))
		}

		return lines
	}

	// Fallback to basic transaction data
	tx := m.transaction
	if tx.Description != "" {
		lines = append(lines, RenderKeyValue("Description", tx.Description))
	}
	lines = append(lines, RenderKeyValue("Amount", RenderMoney(tx.Amount)))
	lines = append(lines, RenderKeyValue("Balance", utils.FormatMoney(tx.Balance, "AUD")))
	if tx.Date != 0 {
		lines = append(lines, RenderKeyValue("Date", utils.FormatUnixDate(tx.Date)))
	}
	if tx.Category != "" {
		lines = append(lines, RenderKeyValue("Category", tx.Category))
	}
	if tx.IsPending {
		lines = append(lines, RenderKeyValue("Status", "Pending"))
	}

	return lines
}

func (m TxnDetailModel) View() string {
	var b strings.Builder

	b.WriteString(RenderHeader(" Transaction Detail ", m.width))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(RenderLoading(m.spinner, "Loading details..."))
		b.WriteString("\n\n")
		return b.String()
	}

	if m.errorMsg != "" {
		b.WriteString(RenderError(m.errorMsg))
		b.WriteString("\n\n")
		return b.String()
	}

	lines := m.buildLines()

	viewHeight := m.height - 6
	if viewHeight < 5 {
		viewHeight = 5
	}

	scrollY := m.scrollY
	if len(lines) > viewHeight {
		maxScroll := len(lines) - viewHeight
		if scrollY > maxScroll {
			scrollY = maxScroll
		}
		end := scrollY + viewHeight
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[scrollY:end]
	}

	for _, line := range lines {
		b.WriteString("  " + line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	return b.String()
}

func (m TxnDetailModel) SetTransaction(tx models.Transaction, accountID string) TxnDetailModel {
	m.transaction = tx
	m.accountID = accountID
	m.detail = nil
	m.rawBody = nil
	m.scrollY = 0
	m.lines = nil
	m.errorMsg = ""
	m.logoASCII = ""
	return m
}

func (m TxnDetailModel) SetDetail(detail *models.TransactionDetail, rawBody []byte) TxnDetailModel {
	m.detail = detail
	m.rawBody = rawBody
	m.loading = false
	m.errorMsg = ""
	m.lines = m.buildLines()
	return m
}

func (m TxnDetailModel) SetCopied(v bool) TxnDetailModel {
	m.copied = v
	return m
}

func (m TxnDetailModel) IsCopied() bool {
	return m.copied
}

func (m TxnDetailModel) SetLogo(ascii string) TxnDetailModel {
	m.logoASCII = ascii
	m.lines = m.buildLines()
	return m
}

func (m TxnDetailModel) SetLoading(loading bool) TxnDetailModel {
	m.loading = loading
	return m
}

func (m TxnDetailModel) SetError(err error) TxnDetailModel {
	if err != nil {
		m.errorMsg = fmt.Sprintf("%v", err)
	} else {
		m.errorMsg = ""
	}
	m.loading = false
	return m
}

func (m TxnDetailModel) SetSize(w, h int) TxnDetailModel {
	m.width = w
	m.height = h
	return m
}

func (m TxnDetailModel) TransactionID() string {
	return m.transaction.TransactionID
}

func (m TxnDetailModel) AccountID() string {
	return m.accountID
}
