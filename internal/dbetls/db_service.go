package dbetls

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/util"
	_ "github.com/go-sql-driver/mysql"
	cron "github.com/netresearch/go-cron"
)

var once sync.Once
var instance *dbService

type DBService interface {
	Init(*core.AppCtx) error
}

type dbService struct {
	logger *slog.Logger
	config *Config
	cron   *cron.Cron
}

func (d *dbService) Init(*core.AppCtx) error {
	d.schedule()
	return nil
}

func (d *dbService) schedule() {
	for _, table := range d.config.Tables {

		d.logger.Info("scheduling table", "database_name", table.DatabaseName, "cron", table.Cron)
		d.cron.AddFunc(table.Cron, func() {
			d.logger.Info("running scheduled task for table", "database_name", table.DatabaseName)
			// Here you would call the function that processes the table
			d.processIngestion(table)
		}, cron.WithName(table.Name))
	}
}

func NewDBService(ctx *core.AppCtx) (DBService, error) {
	ctx.Logger.Info("in dbetls.NewDBService, initializing DBService")
	var reterr error
	once.Do(func() {
		config, err := Init(ctx, "config.xml")
		if err != nil {
			ctx.Logger.Error("failed to initialize config", slog.Any("error", err))
			reterr = err
			return
		}

		instance = &dbService{
			logger: ctx.Logger,
			config: config,
			cron:   ctx.Cron,
		}
	})
	if reterr != nil {
		return nil, fmt.Errorf("failed to create DBService instance: %w", reterr)
	}
	ctx.Logger.Info("DBService instance created successfully")
	return instance, reterr
}

func (d *dbService) processIngestion(table Table) {
	d.logger.Info("processing ingestion for table", "database_name", table.DatabaseName)

	// Step 1: Resolve DB connection details from environment variables.
	// Env vars are expected as: <DATABASE_NAME>_HOST, <DATABASE_NAME>_PORT,
	// <DATABASE_NAME>_USER, <DATABASE_NAME>_PASSWORD, <DATABASE_NAME>_DBNAME
	// dbHost := util.Getenv(table.DatabaseName+"_HOST", "localhost")
	// dbPort := util.Getenv(table.DatabaseName+"_PORT", "3306")
	// dbUser := util.Getenv(table.DatabaseName+"_USER", "")
	// dbPass := util.Getenv(table.DatabaseName+"_PASSWORD", "")
	// dbName := util.Getenv(table.DatabaseName+"_DBNAME", table.DatabaseName)

	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	dsn := util.Getenv(table.DatabaseName, "")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		d.logger.Error("failed to open database connection", "database_name", table.DatabaseName, slog.Any("error", err))
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		d.logger.Error("failed to ping database", "database_name", table.DatabaseName, slog.Any("error", err))
		return
	}

	// Step 2: Execute the query to fetch data.
	rows, err := db.Query(table.Query.Cdata)
	if err != nil {
		d.logger.Error("failed to execute query", "database_name", table.DatabaseName, slog.Any("error", err))
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		d.logger.Error("failed to get columns", "database_name", table.DatabaseName, slog.Any("error", err))
		return
	}

	batchSize := d.parseNonNegativeInt(table.BatchSize)
	maxRecords := d.parseNonNegativeInt(table.MaxRecords)
	waitMS := d.parseNonNegativeInt(table.WaitMS)

	var records []map[string]any
	recordCount := 0

	for rows.Next() {
		if maxRecords >= 0 && (maxRecords > 0 && recordCount >= maxRecords) {
			d.logger.Info("reached max records limit", "max_records", maxRecords, "database_name", table.DatabaseName)
			break
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			d.logger.Error("failed to scan row", "database_name", table.DatabaseName, slog.Any("error", err))
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for JSON serialization
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		// Apply filter mutations to the row.
		d.applyFilterMutations(&row, table.Filter)
		records = append(records, row)
		recordCount++

		// Process in batches if batch size is configured.
		if batchSize > 0 && len(records) >= batchSize {
			d.processOutput(records, table.Output)
			d.logger.Info("processed batch", "batch_size", len(records), "database_name", table.DatabaseName)
			records = nil
			if waitMS > 0 {
				time.Sleep(time.Duration(waitMS) * time.Millisecond)
			}
		}
	}

	if err := rows.Err(); err != nil {
		d.logger.Error("error iterating over rows", "database_name", table.DatabaseName, slog.Any("error", err))
	}

	// Process any remaining records.
	if len(records) > 0 {
		d.processOutput(records, table.Output)
		d.logger.Info("processed final batch", "batch_size", len(records), "database_name", table.DatabaseName)
	}

	d.logger.Info("ingestion completed", "database_name", table.DatabaseName, "total_records", recordCount)
}

// parseNonNegativeInt parses s as a non-negative integer, returning 0 for
// empty, non-numeric, or negative input.
func (d *dbService) parseNonNegativeInt(s string) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return 0
}

// applyFilterMutations applies configured filter transformations to a row.
func (d *dbService) applyFilterMutations(row *map[string]interface{}, filter Filter) {
	mut := filter.Mutate

	// Copy values: copy from one field to another
	for _, cv := range mut.CopyValue {
		if val, ok := (*row)[cv.From]; ok {
			(*row)[cv.To] = val
		}
	}

	// Remove fields
	for _, field := range mut.RemoveFields.Field {
		if mut.RemoveFields.IgnoreCase == "true" {
			// Case-insensitive removal: find matching key
			for key := range *row {
				if strings.EqualFold(key, field) {
					delete(*row, key)
					break
				}
			}
		} else {
			delete(*row, field)
		}
	}

	// Add fields
	for _, f := range mut.AddFields.Field {
		(*row)[f.Key] = f.Value
	}

	// Lowercase fields
	for _, field := range mut.LowercaseFields.Field {
		if val, ok := (*row)[field]; ok {
			if s, ok := val.(string); ok {
				(*row)[field] = strings.ToLower(s)
			}
		}
	}
}

// processOutput routes processed records to the configured output target.
func (d *dbService) processOutput(records []map[string]interface{}, output Output) {
	if output.SkipOutput == "true" {
		d.logger.Info("output skipped", "database_name", output.Name)
		return
	}

	switch strings.ToLower(output.Type) {
	case "api":
		switch strings.ToLower(output.Target) {
		case "elasticsearch", "elastic", "opensearch":
			d.sendToElasticsearch(records, output)
		default:
			d.sendToAPI(records, output)
		}
	case "file":
		d.writeToFile(records, output)
	case "db", "database":
		d.insertToDB(records, output)
	default:
		d.logger.Warn("unknown output target", "target", output.Target, "database_name", output.Name)
	}
}

// doHTTPRequest executes an HTTP request and returns the response body.
// Returns empty byte slice on error (error is already logged).
func (d *dbService) doHTTPRequest(req *http.Request, logMsg string) []byte {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		d.logger.Error(logMsg, "url", req.URL.String(), slog.Any("error", err))
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		d.logger.Error(logMsg+" failed", "status", resp.StatusCode, "body", string(body))
		return nil
	}

	return body
}

// sendToElasticsearch sends records to Elasticsearch using the _bulk API.
func (d *dbService) sendToElasticsearch(records []map[string]interface{}, output Output) {
	var buf bytes.Buffer

	for _, record := range records {
		// Action metadata
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": output.Name,
			},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		// Document source
		docJSON, _ := json.Marshal(record)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	bulkURL := strings.TrimRight(output.URL, "/") + "/_bulk"
	req, err := http.NewRequest("POST", bulkURL, &buf)
	if err != nil {
		d.logger.Error("failed to create elasticsearch request", slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	if output.Auth.User != "" {
		req.SetBasicAuth(output.Auth.User, output.Auth.Password)
	}

	body := d.doHTTPRequest(req, "elasticsearch bulk request")
	if body == nil {
		return
	}

	// Check for errors in the ES bulk response
	var bulkResp struct {
		Errors bool `json:"errors"`
	}
	if err := json.Unmarshal(body, &bulkResp); err == nil && bulkResp.Errors {
		d.logger.Error("elasticsearch bulk response contains errors", "body", string(body))
		return
	}

	d.logger.Info("successfully sent records to elasticsearch", "count", len(records), "index", output.Name)
}

// sendToAPI sends records to a generic HTTP API endpoint.
func (d *dbService) sendToAPI(records []map[string]interface{}, output Output) {
	payload, err := json.Marshal(records)
	if err != nil {
		d.logger.Error("failed to marshal records for API", slog.Any("error", err))
		return
	}

	req, err := http.NewRequest("POST", output.URL, bytes.NewReader(payload))
	if err != nil {
		d.logger.Error("failed to create API request", slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if output.Auth.User != "" {
		req.SetBasicAuth(output.Auth.User, output.Auth.Password)
	}

	d.doHTTPRequest(req, "API request")

	d.logger.Info("successfully sent records to API", "count", len(records), "url", output.URL)
}

// writeToFile writes records as JSON to a file.
func (d *dbService) writeToFile(records []map[string]interface{}, output Output) {
	filePath := output.URL
	if filePath == "" {
		filePath = fmt.Sprintf("output_%s.json", output.Name)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		d.logger.Error("failed to marshal records for file", slog.Any("error", err))
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		d.logger.Error("failed to write file", "path", filePath, slog.Any("error", err))
		return
	}

	d.logger.Info("successfully wrote records to file", "count", len(records), "path", filePath)
}

// insertToDB inserts records into a database using the output URL as the DSN.
func (d *dbService) insertToDB(records []map[string]interface{}, output Output) {
	if len(records) == 0 {
		return
	}

	db, err := sql.Open("mysql", output.URL)
	if err != nil {
		d.logger.Error("failed to open destination database", slog.Any("error", err))
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		d.logger.Error("failed to ping destination database", slog.Any("error", err))
		return
	}

	tableName := output.Name
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
		cols[i] = "`" + k + "`"
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
		d.logger.Error("failed to begin transaction", slog.Any("error", err))
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		d.logger.Error("failed to prepare insert statement", slog.Any("error", err))
		return
	}
	defer stmt.Close()

	for _, record := range records {
		args := make([]interface{}, len(keys))
		for i, k := range keys {
			args[i] = record[k]
		}
		if _, err := stmt.Exec(args...); err != nil {
			d.logger.Error("failed to insert record", slog.Any("error", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		d.logger.Error("failed to commit transaction", slog.Any("error", err))
		return
	}

	d.logger.Info("successfully inserted records into db", "count", len(records), "table", tableName)
}
