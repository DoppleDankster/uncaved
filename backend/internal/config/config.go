package config

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/viper"

	"github.com/DoppleDankster/uncaved/internal/instrumentation"
	"github.com/DoppleDankster/uncaved/internal/server"
	"github.com/DoppleDankster/uncaved/internal/store"
)

type Config struct {
	Instrumentation instrumentation.Config `mapstructure:"instrumentation"`
	Server          server.Config          `mapstructure:"server"`
	DB              store.Config           `mapstructure:"db"`
}

func (c Config) Validate() error {
	var errs []error

	if err := c.Instrumentation.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.DB.Validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Load builds the aggregate config from three layers, lowest precedence first:
// built-in defaults, an optional TOML file, then environment overrides of the
// form UNCAVED_SECTION_KEY (e.g. UNCAVED_SERVER_PORT, UNCAVED_INSTRUMENTATION_LOG_LEVEL).
func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigType("toml")

	setDefaults(v)

	v.SetEnvPrefix("UNCAVED")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return Config{}, fmt.Errorf("config: read %q: %w", path, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// setDefaults registers every config key with a default and
// makes viper aware of the value so that's it's properly unmarshalled.
// Empty string means "no default, but the key exists".
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)

	v.SetDefault("instrumentation.service-name", "uncaved")
	v.SetDefault("instrumentation.service-version", "")
	v.SetDefault("instrumentation.environment", "development")
	v.SetDefault("instrumentation.otlp-endpoint", "")
	v.SetDefault("instrumentation.log-level", "info")
	v.SetDefault("instrumentation.trace-sampling", 1.0)
	v.SetDefault("instrumentation.metric-interval", "60s")

	// all fields expect for db.password can be configured from the config file
	// UNCAVED_DB_PASSWORD must be set in the env
	v.SetDefault("db.host", "")
	v.SetDefault("db.username", "")
	v.SetDefault("db.password", "")
	v.SetDefault("db.database", "")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.auto-migrate", true)
}
