package dbconfig

import (
	"database/sql"
	"fmt"
)

// Config is the configuration for the PostgreSQL database.
type Config struct {
	HostPort       string // HostPort for the database connection.
	User           string // Username for connecting to the database.
	Password       string // Password for the database user.
	Database       string // Name of the database to connect to.
	SSLModeEnabled bool   // SSL mode for the database connection.
}

// ConfigOption defines a function type for configuring options.
type ConfigOption func(*Config)

// WithHostPort is an option to set the HostPort for the database connection.
func WithHostPort(hostport string) ConfigOption {
	return func(cfg *Config) {
		cfg.HostPort = hostport
	}
}

// WithSSLModeEnabled is an option to set the SSL mode for the database connection.
func WithSSLModeEnabled(sslmode bool) ConfigOption {
	return func(cfg *Config) {
		cfg.SSLModeEnabled = sslmode
	}
}

// NewConfig creates a new Config with the provided mandatory values and options.
func NewConfig(user, password, database string, options ...ConfigOption) *Config {
	cfg := &Config{
		HostPort:       "localhost:5432",
		User:           user,
		Password:       password,
		Database:       database,
		SSLModeEnabled: false,
	}
	for _, option := range options { // Apply options.
		option(cfg)
	}
	return cfg
}

// FormatDSN returns a connection string for connect to PostgreSQL.
func (cfg *Config) FormatDSN() string {
	var sslmode = "disable"
	if cfg.SSLModeEnabled {
		sslmode = "enable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.HostPort, cfg.Database, sslmode)
}

// Open opens a database connection with the provided configuration.
func Open(cfg *Config) (*sql.DB, error) {
	return sql.Open("postgres", cfg.FormatDSN())
}
