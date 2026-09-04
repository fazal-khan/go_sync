package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fazal-khan/go_sync/internal/dbetls"
	"log/slog"
)

type FileAdapter struct {
	logger *slog.Logger
}

func NewFileAdapter(logger *slog.Logger) *FileAdapter {
	return &FileAdapter{logger: logger}
}

func (f *FileAdapter) Output(records []map[string]any, cfg dbetls.Output) error {
	filePath := cfg.URL
	if filePath == "" {
		filePath = filepath.Join(".", fmt.Sprintf("output_%s.json", cfg.Name))
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		f.logger.Error("failed to marshal records for file", "error", err)
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		f.logger.Error("failed to write file", "path", filePath, "error", err)
		return err
	}

	f.logger.Info("successfully wrote records to file", "count", len(records), "path", filePath)
	return nil
}