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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Colors shared by both themes
var (
	violet       = lipgloss.Color("#6366f1")
	green        = lipgloss.Color("#10b981")
	red          = lipgloss.Color("#ef4444")
	orange       = lipgloss.Color("#f97316")
	gray         = lipgloss.Color("#6b7280")
	white        = lipgloss.Color("#f9fafb")
	titleStyle   = lipgloss.NewStyle().Foreground(white).Background(violet).Padding(0, 1).Bold(true)
	errTitle     = titleStyle.Background(red)
	successStyle = lipgloss.NewStyle().Foreground(green)
	errorStyle   = lipgloss.NewStyle().Foreground(red)
	timeoutStyle = lipgloss.NewStyle().Foreground(orange)
	mutedStyle   = lipgloss.NewStyle().Foreground(gray)
	modalStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)

	// Adds thousands separators to numbers
	printer = message.NewPrinter(language.English)
)

// A theme only picks the styles, the layout is the same for all of them
type theme struct {
	app, label, value, key lipgloss.Style
}

var (
	minimalTheme = theme{
		app:   lipgloss.NewStyle().Foreground(violet).Bold(true),
		label: lipgloss.NewStyle(),
		value: lipgloss.NewStyle().Bold(true),
		key:   lipgloss.NewStyle().Bold(true),
	}
	modernTheme = theme{
		app:   lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true),
		label: lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa")),
		value: lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true),
		key:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Background(lipgloss.Color("#1f2937")).Padding(0, 1),
	}
)

type QueryResult struct {
	AppName    string
	QueryLabel string
	Result     *db.Result
}

type DashboardModel struct {
	config      *config.Config
	results     []QueryResult
	loading     bool
	lastRefresh time.Time
	filterApp   string
	showErrors  bool
	width       int
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if m.loading {
				break
			}
			m.loading = true
			m.showErrors = false
			return m, m.refreshData()
		case "?":
			m.showErrors = !m.showErrors && countErrors(m.results) > 0
		case "esc":
			m.showErrors = false
		case "t":
			m.toggleTheme()
		}
	case []QueryResult:
		m.results = msg
		m.loading = false
		m.lastRefresh = time.Now()
	}

	return m, nil
}

func (m *DashboardModel) theme() theme {
	if m.config.GetTheme() == "modern" {
		return modernTheme
	}
	return minimalTheme
}

func (m *DashboardModel) toggleTheme() {
	if m.config.GetTheme() == "modern" {
		m.config.Theme = "minimal"
	} else {
		m.config.Theme = "modern"
	}
	// Persist to config file
	_ = m.config.Save()
}

func countErrors(results []QueryResult) int {
	n := 0
	for _, r := range results {
		if r.Result.Error != nil {
			n++
		}
	}
	return n
}

// appNames returns the displayed apps, sorted
func (m *DashboardModel) appNames() []string {
	var names []string
	for name := range m.config.Apps {
		if m.filterApp == "" || name == m.filterApp {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (m *DashboardModel) refreshData() tea.Cmd {
	return func() tea.Msg {
		names := m.appNames()

		// One slot per app so results keep the sorted order
		// whichever database answers first
		perApp := make([][]QueryResult, len(names))
		var wg sync.WaitGroup
		for i, name := range names {
			wg.Add(1)
			go func() {
				defer wg.Done()
				perApp[i] = queryApp(name, m.config.Apps[name])
			}()
		}
		wg.Wait()

		var all []QueryResult
		for _, results := range perApp {
			all = append(all, results...)
		}
		return all
	}
}

func queryApp(appName string, app config.App) []QueryResult {
	conn, err := db.ConnectByType(app.Type, app.Connection)
	if err != nil {
		return []QueryResult{{
			AppName:    appName,
			QueryLabel: "connection",
			Result:     &db.Result{Error: err},
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
			AppName:    appName,
			QueryLabel: label,
			Result:     result,
		})
	}
	return results
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case int, int32, int64:
		return printer.Sprintf("%d", v)
	case float32, float64:
		return printer.Sprintf("%.2f", v)
	case time.Time:
		return v.Format("2006-01-02 15:04")
	case nil:
		return "null"
	default:
		return firstLine(fmt.Sprintf("%v", val))
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
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

func errorMark(err error) string {
	if isTimeoutError(err) {
		return timeoutStyle.Render("⚠")
	}
	return errorStyle.Render("✗")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

func (m *DashboardModel) View() string {
	var b strings.Builder

	if m.showErrors {
		b.WriteString(m.renderErrorModal())
	} else {
		title := "dashmin"
		if m.filterApp != "" {
			title += " · " + m.filterApp
		}
		b.WriteString(titleStyle.Render(title) + "  " + m.renderStatus() + "\n")

		if !m.lastRefresh.IsZero() {
			for _, name := range m.appNames() {
				b.WriteString("\n")
				b.WriteString(m.renderApp(name))
			}
		}

		keys := []string{"r", "refresh", "t", "theme", "q", "quit"}
		if countErrors(m.results) > 0 {
			keys = append(keys, "?", "errors")
		}
		b.WriteString("\n")
		for i := 0; i < len(keys); i += 2 {
			b.WriteString(m.theme().key.Render(keys[i]) + mutedStyle.Render(" "+keys[i+1]+"   "))
		}
	}

	// Cut long lines instead of letting the terminal wrap them
	if m.width > 0 {
		return lipgloss.NewStyle().MaxWidth(m.width).Render(b.String())
	}
	return b.String()
}

func (m *DashboardModel) renderStatus() string {
	if m.lastRefresh.IsZero() {
		return mutedStyle.Render("Loading…")
	}

	total := plural(len(m.results), "query", "queries")
	summary := successStyle.Render("✓ " + total)
	if failed := countErrors(m.results); failed > 0 {
		summary = errorStyle.Render(fmt.Sprintf("✗ %d of %s failed", failed, total))
	}

	updated := "updated " + m.lastRefresh.Format("15:04:05")
	if m.loading {
		updated = "refreshing…"
	}
	return summary + mutedStyle.Render(" · "+updated)
}

func (m *DashboardModel) renderApp(name string) string {
	var rows []QueryResult
	var values []string
	labelWidth, valueWidth := 0, 0
	for _, r := range m.results {
		if r.AppName != name {
			continue
		}
		value := ""
		if r.Result.Error == nil && len(r.Result.Rows) > 0 && len(r.Result.Rows[0]) > 0 {
			value = formatValue(r.Result.Rows[0][0])
		}
		rows = append(rows, r)
		values = append(values, value)
		labelWidth = max(labelWidth, lipgloss.Width(r.QueryLabel))
		valueWidth = max(valueWidth, lipgloss.Width(value))
	}

	var b strings.Builder
	th := m.theme()
	b.WriteString(th.app.Render(name) + " " + mutedStyle.Render(m.config.Apps[name].Type) + "\n")
	if len(rows) == 0 {
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  no queries yet, try: dashmin query add %s <label> <query>", name)) + "\n")
	}

	label := th.label.Width(labelWidth)
	valueStyle := th.value.Width(valueWidth).Align(lipgloss.Right)
	for i, r := range rows {
		var mark, value string
		switch {
		case r.Result.Error != nil:
			mark = errorMark(r.Result.Error)
			value = errorStyle.Render(firstLine(r.Result.Error.Error()))
		case values[i] == "":
			mark = mutedStyle.Render("–")
			value = mutedStyle.Render("no data")
		default:
			mark = successStyle.Render("✓")
			value = valueStyle.Render(values[i])
		}
		fmt.Fprintf(&b, "%s %s  %s\n", mark, label.Render(r.QueryLabel), value)
	}
	return b.String()
}

func (m *DashboardModel) renderErrorModal() string {
	var b strings.Builder

	b.WriteString(errTitle.Render("Error Details"))
	b.WriteString("\n\n")

	for _, result := range m.results {
		if result.Result.Error != nil {
			fmt.Fprintf(&b, "%s %s.%s\n", errorMark(result.Result.Error), result.AppName, result.QueryLabel)
			fmt.Fprintf(&b, "  %s\n\n", result.Result.Error)
		}
	}

	b.WriteString(mutedStyle.Render(fmt.Sprintf("%s · ? or esc to close", plural(countErrors(m.results), "error", "errors"))))

	style := modalStyle
	if m.width > 4 {
		// Border sits outside Width, so leave room for it
		style = style.Width(min(m.width-2, 100))
	}
	return style.Render(b.String())
}

func RunDashboard(cfg *config.Config, filterApp string) error {
	m := NewDashboard(cfg, filterApp)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
