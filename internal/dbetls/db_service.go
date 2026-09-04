package dbetls

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/filter"
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
	logger    *slog.Logger
	config    *Config
	cron      *cron.Cron
	filter    filter.Filter
	outputter Outputter
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
			d.processIngestion(table)
		}, cron.WithName(table.Name))
	}
}

func NewDBService(ctx *core.AppCtx, outputter Outputter) (DBService, error) {
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
			logger:    ctx.Logger,
			config:    config,
			cron:      ctx.Cron,
			filter:    dbetlsFilterToFilter(config.Tables[0].Filter),
			outputter: outputter,
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

	batchSize := util.ParseNonNegativeInt(table.BatchSize)
	maxRecords := util.ParseNonNegativeInt(table.MaxRecords)
	waitMS := util.ParseNonNegativeInt(table.WaitMS)

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
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		row = filter.Apply(row, d.filter)
		records = append(records, row)
		recordCount++

		if batchSize > 0 && len(records) >= batchSize {
			d.outputter.Output(records, table.Output)
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

	if len(records) > 0 {
		d.outputter.Output(records, table.Output)
		d.logger.Info("processed final batch", "batch_size", len(records), "database_name", table.DatabaseName)
	}

	d.logger.Info("ingestion completed", "database_name", table.DatabaseName, "total_records", recordCount)
}

func dbetlsFilterToFilter(f Filter) filter.Filter {
	var copyValues []filter.CopyValue
	for _, cv := range f.Mutate.CopyValue {
		copyValues = append(copyValues, filter.CopyValue{From: cv.From, To: cv.To, In: cv.In})
	}
	var addFields []filter.Field
	for _, fld := range f.Mutate.AddFields.Field {
		addFields = append(addFields, filter.Field{Key: fld.Key, Value: fld.Value})
	}
	return filter.Filter{Mutate: filter.Mutate{
		CopyValue:       copyValues,
		RemoveFields:    filter.RemoveFields{Field: f.Mutate.RemoveFields.Field, IgnoreCase: f.Mutate.RemoveFields.IgnoreCase},
		AddFields:       filter.AddFields{Field: addFields},
		LowercaseFields: filter.LowercaseFields{Field: f.Mutate.LowercaseFields.Field, For: f.Mutate.LowercaseFields.For, CaseSentitive: f.Mutate.LowercaseFields.CaseSentitive},
		Row:             f.Mutate.Row,
	}}
}