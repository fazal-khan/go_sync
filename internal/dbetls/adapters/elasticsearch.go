package adapters

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fazal-khan/go_sync/internal/dbetls"
	"log/slog"
)

type ElasticsearchAdapter struct {
	http    *HTTPClientAdapter
	logger  *slog.Logger
}

func NewElasticsearchAdapter(http *HTTPClientAdapter, logger *slog.Logger) *ElasticsearchAdapter {
	return &ElasticsearchAdapter{
		http:   http,
		logger: logger,
	}
}

func (e *ElasticsearchAdapter) Output(records []map[string]any, cfg dbetls.Output) error {
	if len(records) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, record := range records {
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": cfg.Name,
			},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		docJSON, _ := json.Marshal(record)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	bulkURL := strings.TrimRight(cfg.URL, "/") + "/_bulk"
	req, err := http.NewRequest("POST", bulkURL, &buf)
	if err != nil {
		e.logger.Error("failed to create elasticsearch request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	if cfg.Auth.User != "" {
		req.SetBasicAuth(cfg.Auth.User, cfg.Auth.Password)
	}

	body, err := e.http.Do(req)
	if err != nil {
		return err
	}
	if body == nil {
		return nil
	}

	var bulkResp struct {
		Errors bool `json:"errors"`
	}
	if err := json.Unmarshal(body, &bulkResp); err == nil && bulkResp.Errors {
		e.logger.Error("elasticsearch bulk response contains errors", "body", string(body))
		return nil
	}

	e.logger.Info("successfully sent records to elasticsearch", "count", len(records), "index", cfg.Name)
	return nil
}