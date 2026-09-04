package adapters

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
		w.Write([]byte(`error`))
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

func TestDo_Timeout(t *testing.T) {
	adapter := NewHTTPClientAdapter(newTestLogger())
	adapter.client.Timeout = time.Millisecond
	req, _ := http.NewRequest("GET", "http://127.0.0.1:1", nil)
	body, err := adapter.Do(req)

	if err == nil {
		t.Error("expected timeout error")
	}
	if body != nil {
		t.Errorf("expected nil body on timeout, got %v", body)
	}
}

func TestDo_InvalidURL(t *testing.T) {
	adapter := NewHTTPClientAdapter(newTestLogger())
	req, err := http.NewRequest("GET", "http://[invalid", nil)
	if err == nil {
		body, doErr := adapter.Do(req)
		if doErr == nil {
			t.Errorf("expected error, got body: %v", body)
		}
	}
}

func TestDo_PostWithBody(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readAll(r, t)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	adapter := NewHTTPClientAdapter(newTestLogger())
	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(`{"key":"value"}`))
	body, err := adapter.Do(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("expected body '{\"ok\":true}', got %q", string(body))
	}
	if gotBody != `{"key":"value"}` {
		t.Errorf("expected request body '{\"key\":\"value\"}', got %q", gotBody)
	}
}