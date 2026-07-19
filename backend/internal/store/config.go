package store

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Port     int    `mapstructure:"port"`
	// AutoMigrate runs pending migrations at server startup. Safe for the
	// single-container deploy (no replica races); set false to gate schema
	// changes behind the explicit `migrate up` command instead.
	AutoMigrate bool `mapstructure:"auto-migrate"`
}

func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("store: hostname is required")
	}
	if c.Database == "" {
		return fmt.Errorf("store: database is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("store: port %v must be in ]0,65535]", c.Port)
	}
	if c.Username == "" {
		return fmt.Errorf("store: username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("store: password is required")
	}
	return nil
}

// dsn returns the connection string to the PGDB
func (c Config) dsn() string {
	userInfo := url.UserPassword(c.Username, c.Password)

	pgURL := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		User:   userInfo,
		Path:   c.Database,
	}

	return pgURL.String()
}
