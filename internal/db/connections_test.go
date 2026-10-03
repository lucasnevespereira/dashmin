package db

import (
	"strings"
	"testing"
)

func TestMongoQueryRejectsInvalidFilter(t *testing.T) {
	// No client needed, the filter is checked before any call to the server
	result, err := (&MongoConnection{}).Query(`users.count({status: "active"})`)
	if err != nil {
		t.Fatal(err)
	}
	if result.Error == nil || !strings.Contains(result.Error.Error(), "invalid filter") {
		t.Errorf("want an invalid filter error, got %v", result.Error)
	}
}

func TestConnectSQLite(t *testing.T) {
	conn, err := ConnectSQLite("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	result, err := conn.Query("SELECT 41 + 1")
	if err != nil || result.Error != nil {
		t.Fatal(err, result.Error)
	}
	if got := result.Rows[0][0]; got != int64(42) {
		t.Errorf("got %v, want 42", got)
	}
}
