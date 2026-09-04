package adapters

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"log/slog"
	"os"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

func TestDo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("GET", srv.URL, nil)
	body, err := adapter.Do(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got %q", string(body))
	}
}

func TestDo_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer srv.Close()

	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("GET", srv.URL, nil)
	body, err := adapter.Do(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if body != nil {
		t.Errorf("expected nil body on non-2xx, got %v", body)
	}
}

func TestDo_ConnectionError(t *testing.T) {
	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("GET", "http://127.0.0.1:1", nil)
	body, err := adapter.Do(req)

	if err == nil {
		t.Error("expected connection error")
	}
	if body != nil {
		t.Errorf("expected nil body on error, got %v", body)
	}
}

func TestDo_StatusBoundary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("GET", srv.URL, nil)
	body, err := adapter.Do(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if body != nil {
		t.Errorf("expected nil body on 400, got %v", body)
	}
}

func TestDo_Redirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final"))
	}))
	defer srv.Close()

	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("GET", srv.URL, nil)
	body, err := adapter.Do(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(body) != "final" {
		t.Errorf("expected body 'final', got %q", string(body))
	}
}