package ui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/lucasnevespereira/dashmin/internal/db"
	"golang.org/x/sync/errgroup"
)

// Theme interface for different UI styles
type Theme interface {
	Header() string
	Status(apps, queries int, lastRefresh time.Time) string
	TableHeader() string
	TableRow(app, query, value, updated string, hasError bool, isTimeout bool) string
	Help(keys []string) string
	ErrorModalHeader() string
	ErrorModalError(app, query, err string, isTimeout bool) string
	ErrorModalFooter(count int) string
	Loading(msg string) string
	Error(msg string) string
	EmptyState() string
}

// MinimalTheme - the original minimal theme
type MinimalTheme struct{}

func (t *MinimalTheme) Header() string {
	return titleStyle.Render("dashmin")
}

func (t *MinimalTheme) Status(apps, queries int, lastRefresh time.Time) string {
	appLabel := "app"
	if apps != 1 {
		appLabel = "apps"
	}
	queryLabel := "query"
	if queries != 1 {
		queryLabel = "queries"
	}
	return successStyle.Render(fmt.Sprintf("✓ %d %s, %d %s", apps, appLabel, queries, queryLabel)) +
		mutedStyle.Render(fmt.Sprintf(" • Updated %s", lastRefresh.Format("15:04:05")))
}

func (t *MinimalTheme) TableHeader() string {
	headers := fmt.Sprintf("%-15s %-20s %-15s %s", "APP", "QUERY", "VALUE", "UPDATED")
	return mutedStyle.Render(headers) + "\n" + mutedStyle.Render(strings.Repeat("-", 70))
}

func (t *MinimalTheme) TableRow(app, query, value, updated string, hasError bool, isTimeout bool) string {
	var status string
	var statusColor lipgloss.Style
	if hasError {
		if isTimeout {
			status = "⚠"
			statusColor = timeoutStyle
		} else {
			status = "✗"
			statusColor = errorStyle
		}
	} else {
		status = "✓"
		statusColor = successStyle
	}
	return fmt.Sprintf("%s %-14s %-20s %-15s %s\n",
		statusColor.Render(status),
		app,
		query,
		value,
		mutedStyle.Render(updated))
}

func (t *MinimalTheme) Help(keys []string) string {
	return mutedStyle.Render(strings.Join(keys, ", "))
}

func (t *MinimalTheme) ErrorModalHeader() string {
	return titleStyle.Render("Error Details")
}

func (t *MinimalTheme) ErrorModalError(app, query, err string, isTimeout bool) string {
	var statusColor lipgloss.Style
	if isTimeout {
		statusColor = timeoutStyle
	} else {
		statusColor = errorStyle
	}
	return fmt.Sprintf("%s %s.%s\n  %s\n\n",
		statusColor.Render("✗"),
		app,
		query,
		err)
}

func (t *MinimalTheme) ErrorModalFooter(count int) string {
	return mutedStyle.Render(fmt.Sprintf("%d error(s) found • Press ? to close", count))
}

func (t *MinimalTheme) Loading(msg string) string {
	if msg != "" {
		return mutedStyle.Render(fmt.Sprintf("Querying %s...", msg))
	}
	return mutedStyle.Render("Loading...")
}

func (t *MinimalTheme) Error(msg string) string {
	return errorStyle.Render(fmt.Sprintf("Error: %v", msg))
}

func (t *MinimalTheme) EmptyState() string {
	return "No apps configured.\n\n" +
		"Quick start:\n" +
		"  dashmin app add myapp postgres \"postgres://user:pass@host/db\"\n" +
		"  dashmin query add myapp users \"SELECT COUNT(*) FROM users\"\n" +
		"  dashmin show\n\n"
}

// ModernTheme - the new modern dark theme
type ModernTheme struct{}

func (t *ModernTheme) Header() string {
	// Empty header - status bar acts as the header
	return ""
}

func (t *ModernTheme) Status(apps, queries int, lastRefresh time.Time) string {
	// dashmin badge + app count in green + query count in green + updated time in gray
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#6366f1")).
		Padding(0, 1).
		Bold(true).
		Render("dashmin")

	bullet := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Render("●")

	appLabel := "app"
	if apps != 1 {
		appLabel = "apps"
	}
	queryLabel := "query"
	if queries != 1 {
		queryLabel = "queries"
	}

	appText := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Render(fmt.Sprintf("%d %s", apps, appLabel))
	queryText := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Render(fmt.Sprintf("%d %s", queries, queryLabel))
	updatedText := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(fmt.Sprintf("Updated %s", lastRefresh.Format("15:04:05")))

	return fmt.Sprintf("%s  %s %s, %s %s  ●  %s", badge, bullet, appText, bullet, queryText, updatedText)
}

func (t *ModernTheme) TableHeader() string {
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Bold(true)

	return fmt.Sprintf("%-12s %-25s %-15s %s",
		headerStyle.Render("APP"),
		headerStyle.Render("QUERY"),
		headerStyle.Render("VALUE"),
		headerStyle.Render("UPDATED"))
}

func (t *ModernTheme) TableRow(app, query, value, updated string, hasError bool, isTimeout bool) string {
	// Match colors from reference image
	appStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))     // Light green
	queryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa"))   // Purple
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24"))   // Yellow/gold
	updatedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")) // Gray

	var displayValue string
	if hasError {
		if isTimeout {
			displayValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316")).Render("TIMEOUT")
		} else {
			displayValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render("ERROR")
		}
	} else {
		displayValue = valueStyle.Render(value)
	}

	return fmt.Sprintf("%-12s %-25s %-15s %s\n",
		appStyle.Render(app),
		queryStyle.Render(query),
		displayValue,
		updatedStyle.Render(updated))
}

func (t *ModernTheme) Help(keys []string) string {
	var parts []string
	for _, key := range keys {
		// Parse key format like "r: refresh" or "?: errors"
		keyChar := key[:1]
		keyLabel := key[3:]

		// Use darker background for key badge like in reference image
		keyBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ca3af")).
			Background(lipgloss.Color("#1f2937")).
			Padding(0, 1).
			Render(keyChar)

		keyText := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render(keyLabel)

		parts = append(parts, fmt.Sprintf("%s %s", keyBadge, keyText))
	}
	return strings.Join(parts, "  ")
}

func (t *ModernTheme) ErrorModalHeader() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#dc2626")).
		Padding(0, 1).
		Bold(true).
		Render("Error Details")
}

func (t *ModernTheme) ErrorModalError(app, query, err string, isTimeout bool) string {
	appStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	queryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa"))
	var errColor lipgloss.Style
	if isTimeout {
		errColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316"))
	} else {
		errColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
	}

	return fmt.Sprintf("%s.%s\n  %s\n\n",
		appStyle.Render(app),
		queryStyle.Render(query),
		errColor.Render(err))
}

func (t *ModernTheme) ErrorModalFooter(count int) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(
		fmt.Sprintf("%d error(s) found • Press ? to close", count))
}

func (t *ModernTheme) Loading(msg string) string {
	if msg != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(fmt.Sprintf("Querying %s...", msg))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("Loading...")
}

func (t *ModernTheme) Error(msg string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render(fmt.Sprintf("Error: %v", msg))
}

func (t *ModernTheme) EmptyState() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render(
		"No apps configured.\n\n"+
			"Quick start:\n") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb")).Render(
			"  dashmin app add myapp postgres \"postgres://user:pass@host/db\"\n"+
				"  dashmin query add myapp users \"SELECT COUNT(*) FROM users\"\n"+
				"  dashmin show\n\n")
}

// Minimal color scheme (for backwards compatibility)
var (
	violet       = lipgloss.Color("#6366f1")
	green        = lipgloss.Color("#10b981")
	red          = lipgloss.Color("#ef4444")
	orange       = lipgloss.Color("#f97316")
	gray         = lipgloss.Color("#6b7280")
	white        = lipgloss.Color("#f9fafb")
	titleStyle   = lipgloss.NewStyle().Foreground(white).Background(violet).Padding(0, 1).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(green)
	errorStyle   = lipgloss.NewStyle().Foreground(red)
	timeoutStyle = lipgloss.NewStyle().Foreground(orange)
	mutedStyle   = lipgloss.NewStyle().Foreground(gray)
	modalStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
)

type QueryResult struct {
	AppName     string
	QueryLabel  string
	Result      *db.Result
	LastUpdated time.Time
}

type DashboardModel struct {
	config       *config.Config
	results      []QueryResult
	loading      bool
	lastRefresh  time.Time
	error        error
	filterApp    string
	showErrors   bool
	currentQuery string
	theme        Theme
}

func NewDashboard(cfg *config.Config, filterApp string) *DashboardModel {
	var theme Theme
	if cfg.GetTheme() == "modern" {
		theme = &ModernTheme{}
	} else {
		theme = &MinimalTheme{}
	}

	return &DashboardModel{
		config:    cfg,
		filterApp: filterApp,
		loading:   true,
		theme:     theme,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.refreshData()
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.error = nil
			m.showErrors = false
			return m, m.refreshData()
		case "?":
			if hasErrors(m.results) {
				m.showErrors = !m.showErrors
			}
		case "t":
			m.toggleTheme()
		}
	case []QueryResult:
		m.results = msg
		m.loading = false
		m.lastRefresh = time.Now()
		m.error = nil
		m.currentQuery = ""
	case error:
		m.loading = false
		m.error = msg
	case string:
		// Progress update message
		m.currentQuery = msg
	}

	return m, nil
}

func (m *DashboardModel) toggleTheme() {
	if m.config.GetTheme() == "modern" {
		m.config.Theme = "minimal"
		m.theme = &MinimalTheme{}
	} else {
		m.config.Theme = "modern"
		m.theme = &ModernTheme{}
	}
	// Persist to config file
	_ = m.config.Save()
}

func hasErrors(results []QueryResult) bool {
	for _, r := range results {
		if r.Result.Error != nil {
			return true
		}
	}
	return false
}

func (m *DashboardModel) refreshData() tea.Cmd {
	return func() tea.Msg {
		// Get sorted list of app names for deterministic ordering
		var appNames []string
		for appName := range m.config.Apps {
			if m.filterApp != "" && appName != m.filterApp {
				continue
			}
			appNames = append(appNames, appName)
		}
		sort.Strings(appNames)

		// Use errgroup for concurrent query execution
		g := new(errgroup.Group)
		var mu sync.Mutex
		var allResults []QueryResult

		for _, appName := range appNames {
			app := m.config.Apps[appName]
			g.Go(func() error {
				appResults := queryApp(appName, app)
				mu.Lock()
				allResults = append(allResults, appResults...)
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		return allResults
	}
}

func queryApp(appName string, app config.App) []QueryResult {
	conn, err := db.ConnectByType(app.Type, app.Connection)
	if err != nil {
		return []QueryResult{{
			AppName:     appName,
			QueryLabel:  "Connection",
			Result:      &db.Result{Error: err},
			LastUpdated: time.Now(),
		}}
	}
	defer func() { _ = conn.Close() }()

	// Get sorted list of query labels for deterministic ordering
	var queryLabels []string
	for label := range app.Queries {
		queryLabels = append(queryLabels, label)
	}
	sort.Strings(queryLabels)

	var results []QueryResult
	for _, label := range queryLabels {
		query := app.Queries[label]
		result, err := conn.Query(query)
		if err != nil {
			result = &db.Result{Error: err}
		}

		results = append(results, QueryResult{
			AppName:     appName,
			QueryLabel:  label,
			Result:      result,
			LastUpdated: time.Now(),
		})
	}
	return results
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%.2f", v)
	case string:
		if len(v) > 15 {
			return v[:12] + "..."
		}
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "timed out")
}

func (m *DashboardModel) View() string {
	// If error modal is open, show it
	if m.showErrors {
		return m.renderErrorModal()
	}

	var b strings.Builder

	// Header
	header := m.theme.Header()
	if header != "" {
		b.WriteString(header)
		b.WriteString("\n\n")
	}

	// Status
	if m.loading {
		b.WriteString(m.theme.Loading(m.currentQuery))
		b.WriteString("\n\n")
	} else if m.error != nil {
		b.WriteString(m.theme.Error(m.error.Error()))
		b.WriteString("\n\n")
	} else {
		if len(m.results) == 0 {
			b.WriteString(m.theme.EmptyState())
		} else {
			// Status info
			if m.filterApp != "" {
				b.WriteString(m.theme.Status(1, len(m.results), m.lastRefresh))
			} else {
				b.WriteString(m.theme.Status(len(m.config.Apps), len(m.results), m.lastRefresh))
			}
			b.WriteString("\n\n")

			// Table headers
			b.WriteString(m.theme.TableHeader())
			b.WriteString("\n")

			// Table rows
			for _, result := range m.results {
				var value string
				var hasError bool
				var isTimeout bool

				if result.Result.Error != nil {
					value = "ERROR"
					hasError = true
					isTimeout = isTimeoutError(result.Result.Error)
				} else if len(result.Result.Rows) > 0 && len(result.Result.Rows[0]) > 0 {
					value = formatValue(result.Result.Rows[0][0])
				} else {
					value = "No data"
				}

				b.WriteString(m.theme.TableRow(
					result.AppName,
					result.QueryLabel,
					value,
					result.LastUpdated.Format("15:04:05"),
					hasError,
					isTimeout,
				))
			}
			b.WriteString("\n")
		}
	}

	// Help footer
	if hasErrors(m.results) {
		b.WriteString(m.theme.Help([]string{"r: refresh", "t: theme", "q: quit", "?: errors"}))
	} else {
		b.WriteString(m.theme.Help([]string{"r: refresh", "t: theme", "q: quit"}))
	}

	return b.String()
}

func (m *DashboardModel) renderErrorModal() string {
	var b strings.Builder

	b.WriteString(m.theme.ErrorModalHeader())
	b.WriteString("\n\n")

	var errorCount int
	for _, result := range m.results {
		if result.Result.Error != nil {
			errorCount++
			b.WriteString(m.theme.ErrorModalError(
				result.AppName,
				result.QueryLabel,
				result.Result.Error.Error(),
				isTimeoutError(result.Result.Error),
			))
		}
	}

	b.WriteString(m.theme.ErrorModalFooter(errorCount))

	return modalStyle.Render(b.String())
}

func RunDashboard(cfg *config.Config, filterApp string) error {
	m := NewDashboard(cfg, filterApp)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
