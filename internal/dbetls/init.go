package dbetls

import (
	"encoding/xml"
	"os"

	"github.com/fazal-khan/go_sync/internal/core"
)

func Init(ctx *core.AppCtx, file string) (*Config, error) {
	ctx.Logger.Info("in dbetls.Init, loading config.xml")
	xmlfile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer xmlfile.Close()

	// Parse the XML file and return the config
	var config Config
	err = xml.NewDecoder(xmlfile).Decode(&config)
	if err != nil {
		return nil, err
	}

	ctx.Logger.Info("config.xml loaded successfully")
	ctx.Logger.Info("config details", "database_name", config.Table[0].DatabaseName)

	return &config, nil
}
