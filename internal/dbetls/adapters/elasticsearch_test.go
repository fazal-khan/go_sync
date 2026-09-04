package adapters

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fazal-khan/go_sync/internal/dbetls"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

func TestElasticsearch_Success(t *testing.T) {
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

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}, {"id": "2"}},
		dbetls.Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)

	if gotBody == "" {
		t.Fatal("expected NDJSON body, got empty")
	}
	if countOf(gotBody, '\n') != 4 {
		t.Errorf("expected 4 newlines in NDJSON, got %d", countOf(gotBody, '\n'))
	}
}

func TestElasticsearch_Auth(t *testing.T) {
	var gotUser, gotPass string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "user" || p != "pass" {
			t.Errorf("basic auth mismatch: ok=%v user=%q pass=%q", ok, u, p)
		}
		gotUser, gotPass = u, p
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "elasticsearch", URL: srv.URL, Name: "idx", Auth: dbetls.Auth{User: "user", Password: "pass"}},
	)

	if gotUser != "user" || gotPass != "pass" {
		t.Errorf("auth not set: user=%q pass=%q", gotUser, gotPass)
	}
}

func TestElasticsearch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`boom`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)
}

func TestElasticsearch_ResponseErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":true,"items":[{"index":{"status":400}}]}`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"},
	)
}

func TestElasticsearch_EmptyRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(nil, dbetls.Output{Target: "elasticsearch", URL: srv.URL, Name: "idx"})
}

func TestElasticsearch_DefaultIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors":false}`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewElasticsearchAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "elasticsearch", URL: srv.URL},
	)
}

func readAll(r *http.Request, t *testing.T) string {
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
