package pgtest_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	pgtest "github.com/micheam/go-pgtest"
)

func TestMain(m *testing.M) {
	_ = slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()
	cleanup, err := pgtest.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}

	m.Run()

	if err := cleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to cleanup pg-test: %v\n", err)
	}
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
	require.NoError(t, err)

	// Select
	var got string
	err = db.QueryRow("SELECT id FROM test").Scan(&got)
	require.NoError(t, err)

	// Verify
	if got != id {
		t.Fatalf("got %q, want %q", got, id)
	}
}
