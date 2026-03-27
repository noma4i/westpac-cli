package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noma4i/westpac-cli/api"
	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

type Screen int

const (
	ScreenLoading Screen = iota
	ScreenLogin
	ScreenAccounts
	ScreenTransactions
	ScreenTransactionDetail
	ScreenSummary
)

type AppModel struct {
	width       int
	height      int
	screen      Screen
	prevScreens []Screen

	login     LoginModel
	accounts  AccountsModel
	txns      TransactionsModel
	txnDetail TxnDetailModel
	summary   SummaryModel

	client     *api.Client
	customerID string
	lastEscAt  time.Time
}

// Messages
type AuthSuccessMsg struct {
	CustomerID string
	Result     *api.AuthResult
}

type AuthErrorMsg struct{ Err error }

type AccountsLoadedMsg struct {
	Accounts []models.Account
	Raw      []byte
}

type AccountsErrorMsg struct{ Err error }

type TxnsLoadedMsg struct {
	Transactions []models.Transaction
	Raw          []byte
}

type TxnsErrorMsg struct{ Err error }

type TxnDetailLoadedMsg struct {
	Detail *models.TransactionDetail
	Raw    []byte
}

type TxnDetailErrorMsg struct{ Err error }

type MerchantLogoMsg struct{ ASCII string }

type SummaryLoadedMsg struct {
	NetWorth      *models.NetWorth
	SpendAnalysis *models.SpendAnalysis
	RawNW         []byte
	RawSpend      []byte
	Err           error
}

type SessionRestoredMsg struct {
	Session *models.SessionData
}

type SessionInvalidMsg struct{}

func NewAppModel(client *api.Client) AppModel {
	return AppModel{
		screen:    ScreenLoading,
		client:    client,
		login:     NewLoginModel(),
		accounts:  NewAccountsModel(),
		txns:      NewTransactionsModel(),
		txnDetail: NewTxnDetailModel(),
		summary:   NewSummaryModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.login.spinner.Tick,
		m.tryRestoreSession(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentH := msg.Height - 3 // reserve space for help bar
		m.login = m.login.SetSize(msg.Width, msg.Height)
		m.accounts = m.accounts.SetSize(msg.Width, contentH)
		m.txns = m.txns.SetSize(msg.Width, contentH)
		m.txnDetail = m.txnDetail.SetSize(msg.Width, contentH)
		m.summary = m.summary.SetSize(msg.Width, contentH)
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// Double-Esc to quit from anywhere
		if key == "esc" {
			now := time.Now()
			if now.Sub(m.lastEscAt) < 500*time.Millisecond {
				return m, tea.Quit
			}
			m.lastEscAt = now

			// On Transactions with active search, first Esc clears search
			if m.screen == ScreenTransactions && m.txns.IsSearching() {
				m.txns = m.txns.ClearSearch()
				return m, nil
			}

			// First Esc: go back (if not on top-level screens)
			if m.screen != ScreenLogin && m.screen != ScreenLoading && m.screen != ScreenAccounts {
				return m.goBack()
			}
			return m, nil
		}

		// On Login/Loading screens, only handle esc at app level
		// Everything else goes to the delegate
		if m.screen == ScreenLogin || m.screen == ScreenLoading {
			break
		}

		// On Transactions screen, let all keys through to search
		// except Shift-combos and Tab
		if m.screen == ScreenTransactions {
			switch key {
			case "tab":
				// no-op on transactions
			case "R":
				return m.refresh()
			case "L":
				// no-op, only from accounts/summary
			default:
				break
			}
			break
		}

		switch key {
		case "tab":
			if m.screen == ScreenAccounts {
				m.pushScreen(ScreenSummary)
				m.summary = m.summary.SetLoading(true).SetError(nil)
				return m, tea.Batch(m.summary.spinner.Tick, m.loadSummary())
			} else if m.screen == ScreenSummary {
				return m.goBack()
			}
		case "R":
			return m.refresh()
		case "L":
			if m.screen == ScreenAccounts || m.screen == ScreenSummary {
				api.ClearSession()
				m.screen = ScreenLogin
				m.login = NewLoginModel().SetSize(m.width, m.height)
				m.prevScreens = nil
				return m, nil
			}
		}

	// Session restore
	case SessionRestoredMsg:
		m.customerID = msg.Session.CustomerID
		m.screen = ScreenAccounts
		m.accounts = m.accounts.SetLoading(true)
		return m, tea.Batch(m.accounts.spinner.Tick, m.loadAccounts())

	case SessionInvalidMsg:
		m.screen = ScreenLogin
		return m, nil

	// Auth
	case LoginSubmitMsg:
		m.login = m.login.SetLoading(true)
		return m, tea.Batch(m.login.spinner.Tick, m.doLogin(msg.CustomerID, msg.Password))

	case AuthSuccessMsg:
		m.customerID = msg.CustomerID
		m.login = m.login.SetLoading(false)
		m.screen = ScreenAccounts
		m.accounts = m.accounts.SetLoading(true)
		return m, tea.Batch(m.accounts.spinner.Tick, m.loadAccounts())

	case AuthErrorMsg:
		m.login = m.login.SetLoading(false).SetError(msg.Err)
		return m, nil

	// Accounts
	case AccountsLoadedMsg:
		m.accounts = m.accounts.SetAccounts(msg.Accounts)
		return m, nil

	case AccountsErrorMsg:
		m.accounts = m.accounts.SetError(msg.Err)
		return m, nil

	case AccountSelectedMsg:
		m.pushScreen(ScreenTransactions)
		m.txns = m.txns.SetAccount(msg.Account).SetLoading(true)
		return m, tea.Batch(m.txns.spinner.Tick, m.loadTransactions(msg.Account.AccountID, 0))

	// Transactions
	case TxnsLoadedMsg:
		m.txns = m.txns.SetTransactions(msg.Transactions)
		return m, nil

	case TxnsErrorMsg:
		m.txns = m.txns.SetError(msg.Err)
		return m, nil

	case TxnLoadPageMsg:
		return m, m.loadMoreTransactions(msg.AccountID, msg.Page)

	case TxnPageLoadedMsg:
		m.txns = m.txns.AppendTransactions(msg.Transactions)
		return m, nil

	case TxnSelectedMsg:
		m.pushScreen(ScreenTransactionDetail)
		m.txnDetail = m.txnDetail.SetTransaction(msg.Transaction, msg.AccountID).SetLoading(true)
		return m, tea.Batch(m.txnDetail.spinner.Tick, m.loadTxnDetail(msg.Transaction))

	// Transaction Detail
	case TxnDetailLoadedMsg:
		m.txnDetail = m.txnDetail.SetDetail(msg.Detail, msg.Raw)
		if msg.Detail != nil && msg.Detail.LogoURL != "" {
			return m, loadMerchantLogo(msg.Detail.LogoURL, m.width/3)
		}
		return m, nil

	case MerchantLogoMsg:
		m.txnDetail = m.txnDetail.SetLogo(msg.ASCII)
		return m, nil

	case TxnCopiedMsg:
		m.txnDetail = m.txnDetail.SetCopied(true)
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return TxnCopyResetMsg{}
		})

	case TxnCopyResetMsg:
		m.txnDetail = m.txnDetail.SetCopied(false)
		return m, nil

	case TxnDetailErrorMsg:
		m.txnDetail = m.txnDetail.SetError(msg.Err)
		return m, nil

	// Summary
	case SummaryLoadedMsg:
		if msg.Err != nil {
			m.summary = m.summary.SetError(msg.Err)
		} else {
			m.summary = m.summary.SetError(nil).
				SetNetWorth(msg.NetWorth, msg.RawNW).
				SetSpendAnalysis(msg.SpendAnalysis, msg.RawSpend).
				SetLoading(false)
		}
		return m, nil
	}

	// Delegate to current screen
	var cmd tea.Cmd
	switch m.screen {
	case ScreenLogin:
		m.login, cmd = m.login.Update(msg)
	case ScreenAccounts:
		m.accounts, cmd = m.accounts.Update(msg)
	case ScreenTransactions:
		m.txns, cmd = m.txns.Update(msg)
	case ScreenTransactionDetail:
		m.txnDetail, cmd = m.txnDetail.Update(msg)
	case ScreenSummary:
		m.summary, cmd = m.summary.Update(msg)
	}

	return m, cmd
}

func (m AppModel) View() string {
	var content string
	var help []KeyBinding

	switch m.screen {
	case ScreenLoading:
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			RenderLoading(m.login.spinner, "Checking session..."))
		return content
	case ScreenLogin:
		content = m.login.View()
		help = LoginHelp
	case ScreenAccounts:
		content = m.accounts.View()
		help = AccountsHelp
	case ScreenTransactions:
		content = m.txns.View()
		help = TransactionsHelp
	case ScreenTransactionDetail:
		content = m.txnDetail.View()
		if m.txnDetail.IsCopied() {
			help = TransactionDetailCopiedHelp
		} else {
			help = TransactionDetailHelp
		}
	case ScreenSummary:
		content = m.summary.View()
		help = SummaryHelp
	}

	helpBar := RenderHelp(help, m.width)
	helpHeight := lipgloss.Height(helpBar)

	// Content fills remaining height, help bar at bottom
	contentHeight := m.height - helpHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var styledContent string
	if m.screen == ScreenLogin {
		styledContent = lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, content)
	} else {
		styledContent = lipgloss.NewStyle().Width(m.width).Height(contentHeight).Render(content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, styledContent, helpBar)
}

func (m *AppModel) pushScreen(s Screen) {
	m.prevScreens = append(m.prevScreens, m.screen)
	m.screen = s
}

func (m AppModel) goBack() (AppModel, tea.Cmd) {
	if len(m.prevScreens) > 0 {
		m.screen = m.prevScreens[len(m.prevScreens)-1]
		m.prevScreens = m.prevScreens[:len(m.prevScreens)-1]
	}
	return m, nil
}

func (m AppModel) refresh() (AppModel, tea.Cmd) {
	switch m.screen {
	case ScreenAccounts:
		m.accounts = m.accounts.SetLoading(true)
		return m, tea.Batch(m.accounts.spinner.Tick, m.loadAccounts())
	case ScreenTransactions:
		m.txns = m.txns.SetLoading(true)
		return m, tea.Batch(m.txns.spinner.Tick, m.loadTransactions(m.txns.AccountID(), m.txns.Page()))
	case ScreenSummary:
		m.summary = m.summary.SetLoading(true).SetError(nil)
		return m, tea.Batch(m.summary.spinner.Tick, m.loadSummary())
	}
	return m, nil
}

// Commands (async API calls)

func (m AppModel) tryRestoreSession() tea.Cmd {
	return func() tea.Msg {
		session, err := m.client.LoadSession()
		if err != nil {
			return SessionInvalidMsg{}
		}
		if !m.client.ValidateSession() {
			return SessionInvalidMsg{}
		}
		return SessionRestoredMsg{Session: session}
	}
}

func (m AppModel) doLogin(customerID, password string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.client.Login(customerID, password)
		if err != nil {
			return AuthErrorMsg{Err: err}
		}
		m.client.SaveSession(customerID, result.ProfileID)
		return AuthSuccessMsg{CustomerID: customerID, Result: result}
	}
}

func (m AppModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts, raw, err := m.client.GetAccounts()
		if err != nil {
			return AccountsErrorMsg{Err: err}
		}
		return AccountsLoadedMsg{Accounts: accounts, Raw: raw}
	}
}

func (m AppModel) loadTransactions(accountID string, page int) tea.Cmd {
	return func() tea.Msg {
		txns, raw, err := m.client.GetTransactions(accountID, page, PageSize)
		if err != nil {
			return TxnsErrorMsg{Err: err}
		}
		return TxnsLoadedMsg{Transactions: txns, Raw: raw}
	}
}

func (m AppModel) loadMoreTransactions(accountID string, page int) tea.Cmd {
	return func() tea.Msg {
		txns, _, err := m.client.GetTransactions(accountID, page, PageSize)
		if err != nil {
			return TxnsErrorMsg{Err: err}
		}
		return TxnPageLoadedMsg{Transactions: txns}
	}
}

func (m AppModel) loadTxnDetail(tx models.Transaction) tea.Cmd {
	return func() tea.Msg {
		detail, raw, err := m.client.GetTransactionDetail(tx)
		if err != nil {
			return TxnDetailErrorMsg{Err: err}
		}
		return TxnDetailLoadedMsg{Detail: detail, Raw: raw}
	}
}

func loadMerchantLogo(url string, width int) tea.Cmd {
	return func() tea.Msg {
		ascii, err := utils.RenderImageFromURL(url, width)
		if err != nil {
			return MerchantLogoMsg{}
		}
		return MerchantLogoMsg{ASCII: ascii}
	}
}

func (m AppModel) loadSummary() tea.Cmd {
	return func() tea.Msg {
		var errs []string
		nw, rawNW, err := m.client.GetNetWorth()
		if err != nil {
			errs = append(errs, err.Error())
		}
		sa, rawSpend, err := m.client.GetSpendAnalysis()
		if err != nil {
			errs = append(errs, err.Error())
		}
		msg := SummaryLoadedMsg{
			NetWorth:      nw,
			SpendAnalysis: sa,
			RawNW:         rawNW,
			RawSpend:      rawSpend,
		}
		if len(errs) > 0 {
			msg.Err = fmt.Errorf("%s", fmt.Sprintf("%v", errs))
		}
		return msg
	}
}
