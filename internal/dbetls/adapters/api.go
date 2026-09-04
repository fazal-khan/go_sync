package adapters

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/fazal-khan/go_sync/internal/dbetls"
	"log/slog"
)

type APIAdapter struct {
	http   *HTTPClientAdapter
	logger *slog.Logger
}

func NewAPIAdapter(http *HTTPClientAdapter, logger *slog.Logger) *APIAdapter {
	return &APIAdapter{
		http:   http,
		logger: logger,
	}
}

func (a *APIAdapter) Output(records []map[string]any, cfg dbetls.Output) error {
	payload, err := json.Marshal(records)
	if err != nil {
		a.logger.Error("failed to marshal records for API", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(payload))
	if err != nil {
		a.logger.Error("failed to create API request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if cfg.Auth.User != "" {
		req.SetBasicAuth(cfg.Auth.User, cfg.Auth.Password)
	}

	_, err = a.http.Do(req)
	if err != nil {
		return err
	}

	a.logger.Info("successfully sent records to API", "count", len(records), "url", cfg.URL)
	return nil
}