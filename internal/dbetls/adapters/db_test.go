package adapters

import (
	"testing"

	"github.com/fazal-khan/go_sync/internal/dbetls"
)

func TestDB_InvalidDSN(t *testing.T) {
	adapter := NewDBAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "db", URL: "invalid-dsn", Name: "tbl"},
	)
}

func TestDB_EmptyRecords(t *testing.T) {
	adapter := NewDBAdapter(newTestLogger())
	adapter.Output(nil, dbetls.Output{Target: "db", URL: "invalid-dsn", Name: "tbl"})
}

func TestDB_MarshalError(t *testing.T) {
	adapter := NewDBAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"ch": make(chan int)}},
		dbetls.Output{URL: "invalid-dsn", Name: "tbl"},
	)
}