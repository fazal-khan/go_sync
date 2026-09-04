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

func TestFile_WriteToReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	readOnlyDir := filepath.Join(dir, "readonly")
	os.MkdirAll(readOnlyDir, 0755)
	os.Chmod(readOnlyDir, 0555)
	defer os.Chmod(readOnlyDir, 0755)
	defer os.RemoveAll(dir)

	adapter := NewFileAdapter(newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "file", URL: filepath.Join(readOnlyDir, "out.json"), Name: "tbl"},
	)
}

func TestFile_WriteToNestedPath(t *testing.T) {
	dir := t.TempDir()
	nestedPath := filepath.Join(dir, "nested", "deep", "out.json")

	adapter := NewFileAdapter(newTestLogger())
	os.MkdirAll(filepath.Dir(nestedPath), 0755)
	adapter.Output(
		[]map[string]any{{"id": "1", "name": "test"}},
		dbetls.Output{Target: "file", URL: nestedPath, Name: "tbl"},
	)

	data, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("failed to read nested output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("nested output file is empty")
	}
	if string(data) != "[\n  {\n    \"id\": \"1\",\n    \"name\": \"test\"\n  }\n]" {
		t.Errorf("unexpected content: %s", string(data))
	}
}
