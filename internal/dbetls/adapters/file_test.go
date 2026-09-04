package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fazal-khan/go_sync/internal/dbetls"
)

func TestFile_WriteToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	adapter := NewFileAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "file", URL: path, Name: "tbl"},
	)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestFile_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(originalDir)

	adapter := NewFileAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "file", Name: "mytable"},
	)

	expected := filepath.Join(dir, "output_mytable.json")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected default output file at %s: %v", expected, err)
	}
}

func TestFile_MarshalError(t *testing.T) {
	adapter := NewFileAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"ch": make(chan int)}},
		dbetls.Output{URL: "/tmp/should_not_exist.json", Name: "tbl"},
	)
}