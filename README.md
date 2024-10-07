# go-pgtest

[![Go](https://github.com/micheam/go-pgtest/actions/workflows/go.yml/badge.svg)](https://github.com/micheam/go-pgtest/actions/workflows/go.yml)

`go-pgtest` is a Go module specifically designed to assist in testing applications that depend on PostgreSQL databases. It provides utilities for setting up temporary test databases that are automatically initiated when `go test` is executed and disposed of once the tests are complete. This functionality ensures a clean and isolated testing environment, making it particularly beneficial for developers aiming to streamline and enhance their database testing workflows.

## Usage

To use `go-pgtest`, integrate it into your Go testing environment. The package offers
several functions to help manage and assert database states during tests.

### Functions

- `Start(ctx context.Context) (func() error, error)`: Starts a new database session.
    This function may be called at the beginning of a test to create a new database session.
    We intent to call this from `TestMain` function.

- `Open(t *testing.T, migrationFn func(db *sql.DB) error) *sql.DB`: Opens a new database connection and applies migrations.
    This function may be called within a test to open a new database connection and apply migrations.
    The `migrationFn` parameter is a function that accepts a `*sql.DB` and returns an error.

    *Note*: This may block your process until the database is ready.

## Examples

Here's a basic example of how to use `go-pgtest` in a test:

```go
package mypackage_test

import (
	...

	pgtest "github.com/micheam/go-pgtest"
)

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

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

func TestDatabaseOperations(t *testing.T) {
	// Setup
	migrationfn := func(db *sql.DB) error {
		// Add your migration code here
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS test (id uuid not null primary key)")
		return err
	}
	db := pgtest.Open(t, migrationfn)
	defer db.Close()

	tx := db.MustBegin()
	defer tx.Rollback()

	// Now, you can perform database operations
	_, err := db.Exec("INSERT INTO test (id) VALUES ($1);", uuid.NewString())
	require.NoError(t, err)
}
```

For more detailed examples, please refer to the [pgtest_test.go](pgtest_test.go) file.

## Author

This module was developed by [Michito Maeda](https://micheam.com)
Contributions and feedback are welcome.

## Acknowledgements

`go-pgtest` relies heavily on the `github.com/ory/dockertest` library. We would like to express our gratitude to the maintainers of `dockertest` for their excellent work, which makes this module possible.

