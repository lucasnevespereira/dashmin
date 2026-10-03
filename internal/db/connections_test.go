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
