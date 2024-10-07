# go-pgtest

## Description

`go-pgtest` is a Go module specifically designed to assist in testing applications that depend on PostgreSQL databases. It provides utilities for setting up temporary test databases that are automatically initiated when `go test` is executed and disposed of once the tests are complete. This functionality ensures a clean and isolated testing environment, making it particularly beneficial for developers aiming to streamline and enhance their database testing workflows.

## Usage

To use `go-pgtest`, integrate it into your Go testing environment. The package offers
several functions to help manage and assert database states during tests.

### Functions

- `AssertRecordCount(t *testing.T, tx *sqlx.Tx, expectedCount int, table string, filterQuery string, ...) bool`: Asserts that a table contains a specific number of records matching a filter.
- `AssertRecordExists(t *testing.T, tx *sqlx.Tx, table string, filterQuery string, params ...any) bool`: Asserts that a record exists in a table matching a filter.
- `AssertRecordNotExists(t *testing.T, tx *sqlx.Tx, table string, filterQuery string, params ...any) bool`: Asserts that no record exists in a table matching a filter.
- `DebugEnable()`: Enables debugging output.
- `Open(t *testing.T, migrationFn func(db *sql.DB) error) *sql.DB`: Opens a new database connection and applies migrations.
- `Start(ctx context.Context) (func() error, error)`: Starts a new database session.

## Examples

Here's a basic example of how to use `go-pgtest` in a test:

```go
package mypackage_test

import (
    "testing"
    "github.com/micheam/go-pgtest"
    "github.com/jmoiron/sqlx"
)

func TestDatabaseOperations(t *testing.T) {
    db := go-pgtest.Open(t, func(db *sql.DB) error {
        // Apply migrations
        return nil
    })
    defer db.Close()

    tx := db.MustBegin()
    defer tx.Rollback()

    go-pgtest.AssertRecordCount(t, tx, 1, "users", "WHERE active = ?", true)
}
```

## Author

This module was developed by [Michito Maeda](https://micheam.com)
Contributions and feedback are welcome.

## License

`go-pgtest` is licensed under the MIT License.
See the LICENSE file for details.
