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

// Minimal color scheme
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
}

func NewDashboard(cfg *config.Config, filterApp string) *DashboardModel {
	return &DashboardModel{
		config:    cfg,
		filterApp: filterApp,
		loading:   true,
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

	// Title with color
	if m.filterApp != "" {
		b.WriteString(titleStyle.Render(fmt.Sprintf("dashmin - %s", m.filterApp)))
	} else {
		b.WriteString(titleStyle.Render("dashmin"))
	}
	b.WriteString("\n\n")

	// Status with color
	if m.loading {
		if m.currentQuery != "" {
			b.WriteString(mutedStyle.Render(fmt.Sprintf("Querying %s...", m.currentQuery)))
		} else {
			b.WriteString(mutedStyle.Render("Loading..."))
		}
		b.WriteString("\n\n")
	} else if m.error != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.error)))
		b.WriteString("\n\n")
	} else {
		if len(m.results) == 0 {
			b.WriteString("No apps configured.\n\n")
			b.WriteString("Quick start:\n")
			b.WriteString("  dashmin app add myapp postgres \"postgres://user:pass@host/db\"\n")
			b.WriteString("  dashmin query add myapp users \"SELECT COUNT(*) FROM users\"\n")
			b.WriteString("  dashmin show\n\n")
		} else {
			// Status info with color
			if m.filterApp != "" {
				b.WriteString(successStyle.Render(fmt.Sprintf("✓ %d queries", len(m.results))))
			} else {
				b.WriteString(successStyle.Render(fmt.Sprintf("✓ %d apps, %d queries", len(m.config.Apps), len(m.results))))
			}
			b.WriteString(mutedStyle.Render(fmt.Sprintf(" • Updated %s", m.lastRefresh.Format("15:04:05"))))
			b.WriteString("\n\n")

			// Simple table with colored headers
			headers := fmt.Sprintf("%-15s %-20s %-15s %s", "APP", "QUERY", "VALUE", "UPDATED")
			b.WriteString(mutedStyle.Render(headers))
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render(strings.Repeat("-", 70)))
			b.WriteString("\n")

			for _, result := range m.results {
				var value, status string
				var statusColor lipgloss.Style

				if result.Result.Error != nil {
					value = "ERROR"
					if isTimeoutError(result.Result.Error) {
						status = "⚠"
						statusColor = timeoutStyle
					} else {
						status = "✗"
						statusColor = errorStyle
					}
				} else if len(result.Result.Rows) > 0 && len(result.Result.Rows[0]) > 0 {
					value = formatValue(result.Result.Rows[0][0])
					status = "✓"
					statusColor = successStyle
				} else {
					value = "No data"
					status = "?"
					statusColor = mutedStyle
				}

				b.WriteString(fmt.Sprintf("%s %-14s %-20s %-15s %s\n",
					statusColor.Render(status),
					result.AppName,
					result.QueryLabel,
					value,
					mutedStyle.Render(result.LastUpdated.Format("15:04:05"))))
			}
			b.WriteString("\n")
		}
	}

	// Help with muted color
	if hasErrors(m.results) {
		b.WriteString(mutedStyle.Render("r: refresh, q: quit, ?: errors"))
	} else {
		b.WriteString(mutedStyle.Render("r: refresh, q: quit"))
	}

	return b.String()
}

func (m *DashboardModel) renderErrorModal() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Error Details"))
	b.WriteString("\n\n")

	var errorCount int
	for _, result := range m.results {
		if result.Result.Error != nil {
			errorCount++
			var statusColor lipgloss.Style
			if isTimeoutError(result.Result.Error) {
				statusColor = timeoutStyle
			} else {
				statusColor = errorStyle
			}

			b.WriteString(fmt.Sprintf("%s %s.%s\n", statusColor.Render("✗"), result.AppName, result.QueryLabel))
			b.WriteString(fmt.Sprintf("  %s\n\n", result.Result.Error))
		}
	}

	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d error(s) found • Press ? to close", errorCount)))

	return modalStyle.Render(b.String())
}

func RunDashboard(cfg *config.Config, filterApp string) error {
	m := NewDashboard(cfg, filterApp)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
