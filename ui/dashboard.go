package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasnevespereira/dashmin/config"
	"github.com/lucasnevespereira/dashmin/db"
)

// Minimal color scheme
var (
	violet = lipgloss.Color("#6366f1")
	green  = lipgloss.Color("#10b981")
	red    = lipgloss.Color("#ef4444")
	gray   = lipgloss.Color("#6b7280")
	white  = lipgloss.Color("#f9fafb")

	titleStyle = lipgloss.NewStyle().Foreground(white).Background(violet).Padding(0, 1).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(green)
	errorStyle = lipgloss.NewStyle().Foreground(red)
	mutedStyle = lipgloss.NewStyle().Foreground(gray)
)

type QueryResult struct {
	AppName     string
	QueryLabel  string
	Result      *db.Result
	LastUpdated time.Time
}

type DashboardModel struct {
	config      *config.Config
	results     []QueryResult
	loading     bool
	lastRefresh time.Time
	error       error
	filterApp   string
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
			return m, m.refreshData()
		}
	case []QueryResult:
		m.results = msg
		m.loading = false
		m.lastRefresh = time.Now()
		m.error = nil
	case error:
		m.loading = false
		m.error = msg
	}

	return m, nil
}

func (m *DashboardModel) refreshData() tea.Cmd {
	return func() tea.Msg {
		var results []QueryResult

		for appName, app := range m.config.Apps {
			// Skip apps not matching filter
			if m.filterApp != "" && appName != m.filterApp {
				continue
			}

			// Connect to database
			conn, err := ConnectDatabase(app)
			if err != nil {
				results = append(results, QueryResult{
					AppName:     appName,
					QueryLabel:  "Connection",
					Result:      &db.Result{Error: err},
					LastUpdated: time.Now(),
				})
				continue
			}
			defer conn.Close()

			// Execute queries
			for label, query := range app.Queries {
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
		}

		return results
	}
}

func ConnectDatabase(app config.App) (db.Connection, error) {
	switch app.Type {
	case "postgres":
		return db.ConnectPostgres(app.Connection)
	case "mysql":
		return db.ConnectMySQL(app.Connection)
	case "mongodb":
		return db.ConnectMongoDB(app.Connection)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", app.Type)
	}
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
		return fmt.Sprintf("%v", v)
	}
}

func (m *DashboardModel) View() string {
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
		b.WriteString(mutedStyle.Render("Loading..."))
		b.WriteString("\n\n")
	} else if m.error != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.error)))
		b.WriteString("\n\n")
	} else {
		if len(m.results) == 0 {
			b.WriteString("No apps configured.\n\n")
			b.WriteString("Quick start:\n")
			b.WriteString("  dashmin add myapp postgres \"postgres://user:pass@host/db\"\n")
			b.WriteString("  dashmin all\n\n")
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
					status = "✗"
					statusColor = errorStyle
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
	b.WriteString(mutedStyle.Render("r: refresh, q: quit"))

	return b.String()
}



func RunDashboard(cfg *config.Config, filterApp string) error {
	m := NewDashboard(cfg, filterApp)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}