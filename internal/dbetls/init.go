package dbetls

import (
	"encoding/xml"
	"os"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/util"
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

	// Substitute environment variables in decoded config strings
	substituteConfigEnvVars(&config)

	ctx.Logger.Info("config.xml loaded successfully")
	ctx.Logger.Info("config details", "database_name", config.Table[0].DatabaseName)

	return &config, nil
}

// substituteConfigEnvVars walks known string fields in the Config and replaces
// ${...} placeholders with environment variable values.
func substituteConfigEnvVars(cfg *Config) {
	for i := range cfg.Table {
		t := &cfg.Table[i]

		// Output fields
		t.Output.URL = util.SubstituteEnvVars(t.Output.URL)
		t.Output.Auth.User = util.SubstituteEnvVars(t.Output.Auth.User)
		t.Output.Auth.Password = util.SubstituteEnvVars(t.Output.Auth.Password)
		t.Output.Result.Cdata = util.SubstituteEnvVars(t.Output.Result.Cdata)

		// Query field
		t.Query.Cdata = util.SubstituteEnvVars(t.Query.Cdata)
	}
}
