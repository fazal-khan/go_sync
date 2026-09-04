package dbetls

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestService() *dbService {
	return &dbService{logger: noopLogger()}
}

// --- helpers ---

func TestParseNonNegativeInt(t *testing.T) {
	d := newTestService()

	cases := map[string]int{
		"100":  100,
		"  50 ": 50,
		"1":    1,
		"200":  200,
		"0":    0,
		"-5":   0,
		"-3":   0,
		"-1":   0,
		"abc":  0,
		"x":    0,
		"nope": 0,
		"":     0,
	}
	for in, want := range cases {
		if got := d.parseNonNegativeInt(in); got != want {
			t.Errorf("parseNonNegativeInt(%q) = %d, want %d", in, got, want)
		}
	}
}

// --- applyFilterMutations ---

func TestApplyFilterMutations_All(t *testing.T) {
	d := newTestService()
	row := map[string]interface{}{
		"id":   "42",
		"name": "MiXeD",
		"keep": "yes",
	}
	filter := Filter{
		Mutate: Mutate{
			CopyValue: []CopyValue{{From: "id", To: "_id"}},
			RemoveFields: RemoveFields{
				Field:      []string{"name"},
				IgnoreCase: "",
			},
			AddFields: AddFields{
				Field: []Field{{Key: "type", Value: "A"}},
			},
			LowercaseFields: LowercaseFields{
				Field: []string{"name"},
			},
		},
	}

	d.applyFilterMutations(&row, filter)

	if row["_id"] != "42" {
		t.Errorf("copy-value failed: _id = %v, want 42", row["_id"])
	}
	if _, exists := row["name"]; exists {
		t.Errorf("remove-field failed: name still present")
	}
	if row["type"] != "A" {
		t.Errorf("add-field failed: type = %v, want A", row["type"])
	}
}

func TestApplyFilterMutations_IgnoreCaseRemove(t *testing.T) {
	d := newTestService()
	row := map[string]interface{}{
		"NAME": "value",
		"keep": "yes",
	}
	filter := Filter{
		Mutate: Mutate{
			RemoveFields: RemoveFields{
				Field:      []string{"name"},
				IgnoreCase: "true",
			},
		},
	}

	d.applyFilterMutations(&row, filter)

	if _, exists := row["NAME"]; exists {
		t.Errorf("ignore-case remove failed: NAME still present")
	}
	if row["keep"] != "yes" {
		t.Errorf("unrelated field removed: keep = %v", row["keep"])
	}
}

func TestApplyFilterMutations_LowercaseFields(t *testing.T) {
	d := newTestService()
	row := map[string]interface{}{
		"name": "HeLLo",
	}
	filter := Filter{
		Mutate: Mutate{
			LowercaseFields: LowercaseFields{
				Field: []string{"name"},
			},
		},
	}

	d.applyFilterMutations(&row, filter)

	if row["name"] != "hello" {
		t.Errorf("lowercase failed: name = %v, want hello", row["name"])
	}
}

func TestApplyFilterMutations_CopyMissingFrom(t *testing.T) {
	d := newTestService()
	row := map[string]interface{}{"other": "1"}
	filter := Filter{
		Mutate: Mutate{
			CopyValue: []CopyValue{{From: "missing", To: "target"}},
		},
	}

	d.applyFilterMutations(&row, filter)

	if _, exists := row["target"]; exists {
		t.Errorf("copy from missing field should be a no-op, but target set")
	}
}

// --- processOutput routing ---

func TestProcessOutput_SkipOutput(t *testing.T) {
	d := newTestService()
	// skip-output=true should short-circuit regardless of target
	d.processOutput([]map[string]interface{}{{"a": 1}}, Output{
		Target:     "elasticsearch",
		SkipOutput: "true",
	})
	// No panic, no network call attempted (httptest would fail otherwise)
}

func TestProcessOutput_UnknownTarget(t *testing.T) {
	d := newTestService()
	d.processOutput([]map[string]interface{}{{"a": 1}}, Output{
		Target: "not-a-real-target",
	})
	// Should log warn and return without panicking
}

// --- ES bulk output (httptest) ---

func TestSendToElasticsearch_Success(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/x-ndjson" {
			t.Errorf("unexpected content-type: %q", r.Header.Get("Content-Type"))
		}
		gotBody = readAll(r, t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false,"items":[]}`))
	}))
	defer srv.Close()

	d := newTestService()
	d.sendToElasticsearch(
		[]map[string]interface{}{{"id": "1"}, {"id": "2"}},
		Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)

	if gotBody == "" {
		t.Fatal("expected NDJSON body, got empty")
	}
	// Two action lines + two doc lines = 4 newline-terminated lines
	wantLines := 4
	if countOf(gotBody, '\n') != wantLines {
		t.Errorf("expected %d newlines in NDJSON, got %d", wantLines, countOf(gotBody, '\n'))
	}
}

func TestSendToElasticsearch_Auth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "user" || p != "pass" {
			t.Errorf("basic auth mismatch: ok=%v user=%q pass=%q", ok, u, p)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	d := newTestService()
	d.sendToElasticsearch(
		[]map[string]interface{}{{"id": "1"}},
		Output{Target: "elasticsearch", URL: srv.URL, Name: "idx", Auth: Auth{User: "user", Password: "pass"}},
	)
}

func TestSendToElasticsearch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`boom`))
	}))
	defer srv.Close()

	d := newTestService()
	// Should log error and return, no panic
	d.sendToElasticsearch(
		[]map[string]interface{}{{"id": "1"}},
		Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)
}

func TestSendToElasticsearch_ResponseErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":true,"items":[{"index":{"status":400}}]}`))
	}))
	defer srv.Close()

	d := newTestService()
	// Should log that bulk response contains errors, no panic
	d.sendToElasticsearch(
		[]map[string]interface{}{{"id": "1"}},
		Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)
}

func TestSendToElasticsearch_EmptyRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	d := newTestService()
	// NDJSON with zero records should still send valid empty body without panic
	d.sendToElasticsearch(nil, Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"})
}

// --- API output (httptest) ---

func TestSendToAPI_Success(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		gotBody = readAll(r, t)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := newTestService()
	d.sendToAPI([]map[string]interface{}{{"id": "1"}}, Output{Target: "api", URL: srv.URL})

	if gotBody != `[{"id":"1"}]` {
		t.Errorf("unexpected body: %s", gotBody)
	}
}

func TestSendToAPI_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`bad`))
	}))
	defer srv.Close()

	d := newTestService()
	// Should log error and return, no panic
	d.sendToAPI([]map[string]interface{}{{"id": "1"}}, Output{Target: "api", URL: srv.URL})
}

// --- file output ---

func TestWriteToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	d := newTestService()
	d.writeToFile([]map[string]interface{}{{"id": "1"}}, Output{Target: "file", URL: path, Name: "tbl"})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestWriteToFile_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	// Point cwd at a temp dir so the default output_<name>.json lands there.
	t.Chdir(dir)

	d := newTestService()
	d.writeToFile([]map[string]interface{}{{"id": "1"}}, Output{Target: "file", Name: "mytable"})

	expected := filepath.Join(dir, "output_mytable.json")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected default output file at %s: %v", expected, err)
	}
}

// --- test helpers ---

func readAll(r *http.Request, t *testing.T) string {
	t.Helper()
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 256)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return string(buf)
}

func countOf(s string, r rune) int {
	n := 0
	for _, c := range s {
		if c == r {
			n++
		}
	}
	return n
}

func TestSendToElasticsearch_DefaultIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	d := newTestService()
	// Empty output.Name should not panic and should still send an empty _index
	d.sendToElasticsearch([]map[string]interface{}{{"id": "1"}}, Output{Target: "elasticsearch", URL: srv.URL})
}

// --- processOutput routing ---

func TestProcessOutput_FilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	d := newTestService()
	d.processOutput([]map[string]interface{}{{"id": "1"}}, Output{
		Type: "file",
		URL:  path,
		Name: "tbl",
	})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestProcessOutput_DBPath(t *testing.T) {
	d := newTestService()
	// "db" type with invalid DSN should log error and return without panic
	d.processOutput([]map[string]interface{}{{"id": "1"}}, Output{
		Type: "db",
		URL:  "invalid-dsn",
		Name: "tbl",
	})
}

func TestProcessOutput_DefaultCase(t *testing.T) {
	d := newTestService()
	// Unknown type should log warning and return without panic
	d.processOutput([]map[string]interface{}{{"id": "1"}}, Output{
		Type: "unknown-type",
		Target: "unknown-target",
		Name: "tbl",
	})
}

// --- sendToAPI additional coverage ---

func TestSendToAPI_Auth(t *testing.T) {
	var gotAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && u == "user" && p == "pass" {
			gotAuth = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := newTestService()
	d.sendToAPI([]map[string]interface{}{{"id": "1"}}, Output{
		Target: "api",
		URL:    srv.URL,
		Auth:   Auth{User: "user", Password: "pass"},
	})

	if !gotAuth {
		t.Error("expected basic auth to be set")
	}
}

// --- writeToFile error paths ---

func TestWriteToFile_MarshalError(t *testing.T) {
	d := newTestService()
	// channels cannot be marshaled to JSON — should log error and return
	bad := map[string]interface{}{"ch": make(chan int)}
	d.writeToFile([]map[string]interface{}{bad}, Output{URL: "/tmp/should_not_exist.json", Name: "tbl"})
	// No panic, no file created
}

// --- doHTTPRequest error paths ---

func TestDoHTTPRequest_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`server error`))
	}))
	defer srv.Close()

	d := newTestService()
	req, _ := http.NewRequest("POST", srv.URL, nil)
	body := d.doHTTPRequest(req, "test request")

	if body != nil {
		t.Errorf("expected nil body on error, got %v", body)
	}
}

func TestDoHTTPRequest_ConnectionError(t *testing.T) {
	d := newTestService()
	req, _ := http.NewRequest("POST", "http://127.0.0.1:1", nil)
	body := d.doHTTPRequest(req, "test request")

	if body != nil {
		t.Errorf("expected nil body on connection error, got %v", body)
	}
}
