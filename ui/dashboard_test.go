package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/lucasnevespereira/dashmin/internal/db"
)

func TestFormatValue(t *testing.T) {
	cases := map[string]interface{}{
		"12,847":   int64(12847),
		"9,341.55": 9341.55,
		"89":       89,
		"hello":    "hello\nworld",
		"null":     nil,
	}
	for want, in := range cases {
		if got := formatValue(in); got != want {
			t.Errorf("formatValue(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestView(t *testing.T) {
	value := func(v interface{}) *db.Result { return &db.Result{Rows: [][]interface{}{{v}}} }
	m := NewDashboard(&config.Config{Apps: map[string]config.App{
		"production": {Type: "postgres"},
		"staging":    {Type: "sqlite"},
		"empty":      {Type: "mysql"},
	}}, "")
	m.Update([]QueryResult{
		{"production", "total_users", value(int64(12847))},
		{"production", "revenue_today", value(9341.55)},
		{"production", "a_rather_long_query_label", value(int64(1))},
		{"production", "last_signup", value(time.Date(2026, 10, 3, 9, 30, 0, 0, time.UTC))},
		{"production", "nothing", &db.Result{}},
		{"staging", "broken", &db.Result{Error: errors.New(`relation "orders" does not exist`)}},
		{"staging", "slow", &db.Result{Error: errors.New("query timed out after 10s")}},
	})

	// Set the theme directly, the "t" key would write to the real config file
	for _, name := range []string{"minimal", "modern"} {
		m.config.Theme = name
		view := m.View()
		t.Log("\n" + view)
		for _, want := range []string{"2 of 7 queries failed", "12,847", "no data", "does not exist", "errors", "no queries yet"} {
			if !strings.Contains(view, want) {
				t.Errorf("%s view is missing %q", name, want)
			}
		}
	}

	m.width = 40
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	modal := m.View()
	t.Log("\n" + modal)
	for _, line := range strings.Split(modal, "\n") {
		if len([]rune(line)) > 40 {
			t.Errorf("modal line wider than terminal: %q", line)
		}
	}
}
