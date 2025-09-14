package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasnevespereira/dashmin/config"
	"github.com/lucasnevespereira/dashmin/db"
)

// Modern, minimalist color scheme
var (
	// Brand colors
	primaryColor   = lipgloss.Color("#6366f1")   // Modern indigo
	successColor   = lipgloss.Color("#10b981")   // Emerald
	errorColor     = lipgloss.Color("#ef4444")   // Red
	warningColor   = lipgloss.Color("#f59e0b")   // Amber
	mutedColor     = lipgloss.Color("#6b7280")   // Gray
	textColor      = lipgloss.Color("#f9fafb")   // Light gray
	bgColor        = lipgloss.Color("#111827")   // Dark gray

	// Styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(textColor).
		Background(primaryColor).
		Padding(0, 2).
		MarginBottom(2)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		MarginTop(1)

	appStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(textColor).
		Background(primaryColor).
		Padding(0, 1).
		MarginRight(1)

	queryStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	valueStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)

	tableStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		MarginBottom(1).
		MarginTop(1)

	detailBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(1).
		MarginTop(1)
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
	table       table.Model
	loading     bool
	lastRefresh time.Time
	error       error
	filterApp   string
}

func NewDashboard(cfg *config.Config, filterApp string) *DashboardModel {
	columns := []table.Column{
		{Title: "App", Width: 18},
		{Title: "Query", Width: 25},
		{Title: "Value", Width: 18},
		{Title: "Updated", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(12),
		table.WithFocused(true),
	)

	// Modern table styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor)

	s.Selected = s.Selected.
		Foreground(textColor).
		Background(primaryColor).
		Bold(true)

	s.Cell = s.Cell.
		Foreground(textColor)

	t.SetStyles(s)

	return &DashboardModel{
		config:    cfg,
		table:     t,
		filterApp: filterApp,
		loading:   true,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.refreshData()
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "r", "ctrl+r":
			m.loading = true
			m.error = nil
			return m, m.refreshData()
		}
	case []QueryResult:
		m.results = msg
		m.loading = false
		m.lastRefresh = time.Now()
		m.error = nil
		m.updateTable()
	case error:
		m.loading = false
		m.error = msg
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
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

func (m *DashboardModel) updateTable() {
	var rows []table.Row

	for _, result := range m.results {
		var value string
		var status string

		if result.Result.Error != nil {
			// Show more meaningful error messages in table
			errStr := result.Result.Error.Error()
			if strings.Contains(errStr, "connect") || strings.Contains(errStr, "ping") {
				value = "Connection Failed"
			} else if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "password") {
				value = "Auth Failed"
			} else if len(errStr) > 20 {
				value = errStr[:17] + "..."
			} else {
				value = errStr
			}
			status = "🔴"
		} else if len(result.Result.Rows) > 0 && len(result.Result.Rows[0]) > 0 {
			value = formatValue(result.Result.Rows[0][0])
			status = "🟢"
		} else {
			value = "No data"
			status = "🟡"
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%s %s", status, result.AppName),
			result.QueryLabel,
			value,
			result.LastUpdated.Format("15:04:05"),
		})
	}

	m.table.SetRows(rows)
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

	// Header
	title := "DASHMIN DASHBOARD"
	if m.filterApp != "" {
		title = fmt.Sprintf("DASHMIN - %s", strings.ToUpper(m.filterApp))
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Status indicator
	if m.loading {
		b.WriteString(headerStyle.Render("🔄 Refreshing data..."))
	} else if m.error != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("❌ Error: %v", m.error)))
	} else {
		var statusText string
		if m.filterApp != "" {
			statusText = fmt.Sprintf("✅ %d queries • Updated %s",
				len(m.results),
				m.lastRefresh.Format("15:04:05"))
		} else {
			statusText = fmt.Sprintf("✅ %d apps • %d queries • Updated %s",
				len(m.config.Apps),
				len(m.results),
				m.lastRefresh.Format("15:04:05"))
		}
		b.WriteString(headerStyle.Render(statusText))
	}
	b.WriteString("\n")

	// Main table
	if len(m.results) == 0 && !m.loading {
		b.WriteString(m.renderEmptyState())
	} else {
		b.WriteString(tableStyle.Render(m.table.View()))

		// Detail view
		if len(m.results) > 0 {
			b.WriteString(m.renderDetailView())
		}
	}

	// Help
	b.WriteString(helpStyle.Render("r: refresh • ↑/↓: navigate • q: quit"))

	return b.String()
}

func (m *DashboardModel) renderEmptyState() string {
	var b strings.Builder
	
	emptyStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(2).
		Align(lipgloss.Center)

	content := "No apps configured yet!\n\n" +
		"Quick start:\n" +
		"  dashmin add myapp postgres \"postgres://readonly:password@localhost:5432/myapp?sslmode=disable\"\n" +
		"  dashmin query myapp users \"SELECT COUNT(*) FROM users\"\n" +
		"  dashmin all"

	b.WriteString(emptyStyle.Render(content))
	return b.String()
}

func (m *DashboardModel) renderDetailView() string {
	selected := m.table.Cursor()
	if selected >= len(m.results) {
		return ""
	}

	result := m.results[selected]
	var b strings.Builder

	// App and query info with better spacing
	b.WriteString(fmt.Sprintf("%s %s\n\n", 
		appStyle.Render("📱 "+result.AppName),
		queryStyle.Render("🔍 "+result.QueryLabel)))

	if result.Result.Error != nil {
		// Show full error message with word wrapping
		errorText := fmt.Sprintf("❌ Error: %v", result.Result.Error)
		
		// Break long error messages into multiple lines
		const maxLineLength = 80
		words := strings.Fields(errorText)
		var lines []string
		var currentLine string
		
		for _, word := range words {
			if len(currentLine)+len(word)+1 <= maxLineLength {
				if currentLine == "" {
					currentLine = word
				} else {
					currentLine += " " + word
				}
			} else {
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = word
			}
		}
		if currentLine != "" {
			lines = append(lines, currentLine)
		}
		
		for _, line := range lines {
			b.WriteString(errorStyle.Render(line))
			b.WriteString("\n")
		}
		
		// Add troubleshooting hint for connection errors
		if strings.Contains(result.Result.Error.Error(), "connect") {
			b.WriteString("\n💡 Troubleshooting:\n")
			b.WriteString("• Check if the database is running\n")
			b.WriteString("• Verify connection string format\n")
			b.WriteString("• Check network connectivity\n")
			b.WriteString("• Ensure credentials are correct\n")
		}
	} else {
		// Results table with better formatting
		if len(result.Result.Columns) > 0 && len(result.Result.Rows) > 0 {
			b.WriteString("📋 Results:\n\n")
			
			// Headers
			headerLine := ""
			for _, col := range result.Result.Columns {
				headerLine += fmt.Sprintf("%-20s", col)
			}
			b.WriteString(headerLine + "\n")
			b.WriteString(strings.Repeat("─", len(headerLine)) + "\n")

			// Data (show up to 5 rows)
			maxRows := 5
			for i, row := range result.Result.Rows {
				if i >= maxRows {
					b.WriteString(fmt.Sprintf("... and %d more rows\n", len(result.Result.Rows)-maxRows))
					break
				}
				for _, val := range row {
					b.WriteString(fmt.Sprintf("%-20v", val))
				}
				b.WriteString("\n")
			}
		}
	}

	return detailBoxStyle.Render(b.String())
}


func RunDashboard(cfg *config.Config, filterApp string) error {
	m := NewDashboard(cfg, filterApp)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}