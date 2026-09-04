package adapters

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/fazal-khan/go_sync/internal/dbetls"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
)

type DBAdapter struct {
	logger *slog.Logger
}

func NewDBAdapter(logger *slog.Logger) *DBAdapter {
	return &DBAdapter{logger: logger}
}

func (d *DBAdapter) Output(records []map[string]any, cfg dbetls.Output) error {
	if len(records) == 0 {
		return nil
	}

	db, err := sql.Open("mysql", cfg.URL)
	if err != nil {
		d.logger.Error("failed to open destination database", "error", err)
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		d.logger.Error("failed to ping destination database", "error", err)
		return err
	}

	tableName := cfg.Name
	if tableName == "" {
		tableName = "records"
	}

	var keys []string
	for k := range records[0] {
		keys = append(keys, k)
	}

	cols := make([]string, len(keys))
	placeholders := make([]string, len(keys))
	for i, k := range keys {
		cols[i] = fmt.Sprintf("`%s`", k)
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	tx, err := db.Begin()
	if err != nil {
		d.logger.Error("failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		d.logger.Error("failed to prepare insert statement", "error", err)
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		args := make([]any, len(keys))
		for i, k := range keys {
			args[i] = record[k]
		}
		if _, err := stmt.Exec(args...); err != nil {
			d.logger.Error("failed to insert record", "error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		d.logger.Error("failed to commit transaction", "error", err)
		return err
	}

	d.logger.Info("successfully inserted records into db", "count", len(records), "table", tableName)
	return nil
}
