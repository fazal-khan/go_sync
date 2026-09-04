package adapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazal-khan/go_sync/internal/dbetls"
)

func TestAPI_Success(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type: %q", r.Header.Get("Content-Type"))
		}
		gotBody = readAll(r, t)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewAPIAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "api", URL: srv.URL},
	)

	if gotBody != `[{"id":"1"}]` {
		t.Errorf("unexpected body: %s", gotBody)
	}
}

func TestAPI_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`bad`))
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewAPIAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "api", URL: srv.URL},
	)
}

func TestAPI_Auth(t *testing.T) {
	var gotAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && u == "user" && p == "pass" {
			gotAuth = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewAPIAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"id": "1"}},
		dbetls.Output{Target: "api", URL: srv.URL, Auth: dbetls.Auth{User: "user", Password: "pass"}},
	)

	if !gotAuth {
		t.Error("expected basic auth to be set")
	}
}

func TestAPI_MarshalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server")
	}))
	defer srv.Close()

	httpAdapter := NewHTTPClientAdapter(newTestLogger())
	adapter := NewAPIAdapter(httpAdapter, newTestLogger())
	adapter.Output(
		[]map[string]any{{"ch": make(chan int)}},
		dbetls.Output{Target: "api", URL: srv.URL},
	)
}
