package dbetls

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/fazal-khan/go_sync/internal/core"
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
	for _, table := range d.config.Table {

		d.logger.Info("scheduling table", "database_name", table.DatabaseName, "cron", table.Cron)
		d.cron.AddFunc(table.Cron, func() {
			d.logger.Info("running scheduled task for table", "database_name", table.DatabaseName)
			// Here you would call the function that processes the table
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
