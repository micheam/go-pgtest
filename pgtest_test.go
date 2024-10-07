package pgtest_test

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	pgtest "github.com/micheam/go-pgtest"
)

func TestMain(m *testing.M) {
	_ = slog.SetLogLoggerLevel(slog.LevelDebug)

	// setup
	cleanup, err := pgtest.Start(context.Background())
	if err != nil {
		log.Fatalf("could not start pgtest server: %s", err)
	}

	code := m.Run()

	if err := cleanup(); err != nil {
		log.Fatalf("could not stop pgtest server: %s", err)
	}

	os.Exit(code)
}

func Test_RunPG(t *testing.T) {

	// Setup
	migrationfn := func(db *sql.DB) error {
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS test (id uuid not null primary key)")
		return err
	}
	db := pgtest.Open(t, migrationfn)
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("could not ping database: %s", err)
	}

	// Insert
	id := uuid.NewString()
	_, err := db.Exec("INSERT INTO test (id) VALUES ($1);", id)
	if err != nil {
		t.Fatalf("could not insert: %s", err)
	}

	// Select
	var got string
	err = db.QueryRow("SELECT id FROM test").Scan(&got)
	if err != nil {
		t.Fatalf("could not select: %s", err)
	}

	// Verify
	if got != id {
		t.Fatalf("got %q, want %q", got, id)
	}
}
