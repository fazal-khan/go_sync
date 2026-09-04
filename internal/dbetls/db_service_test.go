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