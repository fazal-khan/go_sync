package dbetls

import (
	"testing"

	"github.com/fazal-khan/go_sync/internal/filter"
)

type mockOutputter struct {
	calls []outputCall
}

type outputCall struct {
	records []map[string]any
	cfg     Output
}

func (m *mockOutputter) Output(records []map[string]any, cfg Output) error {
	m.calls = append(m.calls, outputCall{records: records, cfg: cfg})
	return nil
}

func newTestService(outputter Outputter, f filter.Filter) *dbService {
	return &dbService{
		logger:    noopLogger(),
		config:    &Config{Tables: []Table{{Name: "test", Cron: "* * * * *"}}},
		filter:    f,
		outputter: outputter,
	}
}

func TestNewDBService_Struct(t *testing.T) {
	mock := &mockOutputter{}
	f := filter.Filter{}
	d := newTestService(mock, f)

	if d.outputter != mock {
		t.Error("outputter not set correctly")
	}
	if d.filter.Mutate.Row != "" {
		t.Error("filter should be empty")
	}
}

func TestDBService_ImplementsInterface(t *testing.T) {
	mock := &mockOutputter{}
	f := filter.Filter{}
	d := newTestService(mock, f)
	_ = DBService(d)
}

func TestDBService_DbetlsFilterToFilter(t *testing.T) {
	dbFilter := Filter{
		Mutate: Mutate{
			Row: "_row",
			CopyValue: []CopyValue{{From: "id", To: "_id", In: "root"}},
			RemoveFields: RemoveFields{Field: []string{"name"}, IgnoreCase: "true"},
			AddFields: AddFields{Field: []Field{{Key: "type", Value: "A"}}},
			LowercaseFields: LowercaseFields{Field: []string{"name"}, For: "value"},
		},
	}
	result := dbetlsFilterToFilter(dbFilter)

	if result.Mutate.Row != "_row" {
		t.Errorf("Row = %q, want _row", result.Mutate.Row)
	}
	if len(result.Mutate.CopyValue) != 1 || result.Mutate.CopyValue[0].From != "id" {
		t.Error("CopyValue not converted correctly")
	}
	if len(result.Mutate.RemoveFields.Field) != 1 || result.Mutate.RemoveFields.Field[0] != "name" {
		t.Error("RemoveFields not converted correctly")
	}
	if len(result.Mutate.AddFields.Field) != 1 || result.Mutate.AddFields.Field[0].Key != "type" {
		t.Error("AddFields not converted correctly")
	}
	if len(result.Mutate.LowercaseFields.Field) != 1 || result.Mutate.LowercaseFields.Field[0] != "name" {
		t.Error("LowercaseFields not converted correctly")
	}
}

func TestProcessIngestion_InvalidDSN(t *testing.T) {
	mock := &mockOutputter{}
	f := filter.Filter{}
	d := newTestService(mock, f)

	table := Table{
		DatabaseName: "testdb",
		Query:        Query{Cdata: "SELECT 1"},
		BatchSize:    "0",
		MaxRecords:   "-1",
		WaitMS:       "0",
	}

	d.processIngestion(table)

	if len(mock.calls) != 0 {
		t.Errorf("expected 0 output calls on invalid DSN, got %d", len(mock.calls))
	}
}

