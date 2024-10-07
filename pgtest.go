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

// maxWait is the maximum time to wait for MinIO server to start.
// Adjust this value if you see "context deadline exceeded" error.
//
// This value will be used on the first call to Start().
// If you want to change this value, change it before calling [Start].
var maxWait = 10 * time.Second

var db *sql.DB

var (
	once    sync.Once
	cleanup func() error

	hostAndPort string
	databaseUrl string

	dbpassword = "secret"
	dbuser     = "test_user"
	dbname     = "test_db"

	opt = &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_PASSWORD=" + dbpassword,
			"POSTGRES_USER=" + dbuser,
			"POSTGRES_DB=" + dbname,
			"listen_addresses = '*'",
		},
	}
)

// Start starts a postgresql server in docker.
// It returns a cleanup function to stop the server.
func Start(ctx context.Context) (func() error, error) {
	logger := slog.With("module", "pgtest")

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

		resource, err := pool.RunWithOptions(opt, func(config *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})

		if err != nil {
			return
		}

		hostAndPort = resource.GetHostPort("5432/tcp")

		databaseUrl = dbconfig.NewConfig(dbuser, dbpassword, dbname,
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
