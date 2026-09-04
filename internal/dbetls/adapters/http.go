package adapters

import (
	"io"
	"net/http"
	"time"

	"log/slog"
)

type HTTPClientAdapter struct {
	client  *http.Client
	logger  *slog.Logger
}

func NewHTTPClientAdapter(logger *slog.Logger) *HTTPClientAdapter {
	return &HTTPClientAdapter{
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

func (h *HTTPClientAdapter) Do(req *http.Request) ([]byte, error) {
	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.Error("HTTP request failed", "url", req.URL.String(), "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read response body", "url", req.URL.String(), "error", err)
		return nil, err
	}

	if resp.StatusCode >= 300 {
		h.logger.Error("HTTP request returned non-2xx", "status", resp.StatusCode, "body", string(body))
		return nil, nil
	}

	return body, nil
}