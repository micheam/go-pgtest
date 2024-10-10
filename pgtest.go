package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/micheam/go-pgtest/internal/dbconfig"
)

const (
	DefaultImageTag = "15"
	DefaultDatabase = "test_db"
	DefaultUser     = "test_user"
	DefaultPassword = "secret"
)

type Option func(*Config)

type Config struct {
	imageTag   string
	dbpassword string
	dbuser     string
	dbname     string
}

func WithImageTag(tag string) Option {
	return func(c *Config) {
		c.imageTag = tag
	}
}

func WithDatabase(name string) Option {
	return func(c *Config) {
		c.dbname = name
	}
}

func WithUser(name string) Option {
	return func(c *Config) {
		c.dbuser = name
	}
}

func WithPassword(password string) Option {
	return func(c *Config) {
		c.dbpassword = password
	}
}

// maxWait is the maximum time to wait for MinIO server to start.
// Adjust this value if you see "context deadline exceeded" error.
//
// This value will be used on the first call to Start().
// If you want to change this value, change it before calling [Start].
var maxWait = 10 * time.Second

var (
	once        sync.Once
	cleanup     func() error
	db          *sql.DB
	hostAndPort string
	databaseUrl string
)

// Start starts a postgresql server in docker.
// It returns a cleanup function to stop the server.
func Start(ctx context.Context, opts ...Option) (func() error, error) {
	logger := slog.With("module", "pgtest")

	config := &Config{
		imageTag:   DefaultImageTag,
		dbpassword: DefaultPassword,
		dbuser:     DefaultUser,
		dbname:     DefaultDatabase,
	}
	for _, opt := range opts {
		opt(config)
	}

	var err error
	once.Do(func() {
		var pool *dockertest.Pool
		if pool, err = dockertest.NewPool(""); err != nil {
			err = fmt.Errorf("could not construct pool: %s", err)
			return
		}
		pool.MaxWait = maxWait

		if err = pool.Client.Ping(); err != nil {
			err = fmt.Errorf("could not connect to Docker: %w", err)
			return
		}

		runOpts := &dockertest.RunOptions{
			Repository: "postgres",
			Tag:        config.imageTag,
			Env: []string{
				"POSTGRES_PASSWORD=" + config.dbpassword,
				"POSTGRES_USER=" + config.dbuser,
				"POSTGRES_DB=" + config.dbname,
				"listen_addresses = '*'",
			},
		}
		resource, err := pool.RunWithOptions(runOpts, func(c *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			c.AutoRemove = true
			c.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})

		if err != nil {
			return
		}

		hostAndPort = resource.GetHostPort("5432/tcp")

		databaseUrl = dbconfig.NewConfig(config.dbuser, config.dbpassword, config.dbname,
			dbconfig.WithHostPort(hostAndPort),
			dbconfig.WithSSLModeEnabled(false),
		).FormatDSN()
		logger.DebugContext(ctx, "Connecting to database on url: "+databaseUrl)

		resource.Expire(120) // Tell docker to hard kill the container in 120 seconds
		pool.MaxWait = 120 * time.Second
		if err = pool.Retry(func() error {
			db, err = sql.Open("postgres", databaseUrl)
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			logger.ErrorContext(ctx, "Could not connect to docker", "error", err)
			return
		}

		cleanup = func() error {
			if pool != nil {
				return pool.Purge(resource)
			}
			return nil
		}
	})
	if err != nil {
		return nil, fmt.Errorf("[pgtest] Start: %w", err)
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait
	backoff.Retry(func() error {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}, bo)

	return cleanup, nil
}

// Open opens a database connection.
//
// This function is useful for testing.
// t.Fatal is called if it fails to open a connection.
func Open(t *testing.T, migrationFn func(db *sql.DB) error) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}

	// Wait for the database to be ready.
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait
	er := backoff.Retry(func() error {
		return db.Ping()
	}, bo)
	if er != nil {
		t.Fatalf("Open: %v", er)
	}

	if migrationFn != nil {
		if err := migrationFn(db); err != nil {
			t.Fatalf("migrationFn: %v", err)
		}
	}
	return db
}
