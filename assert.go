package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
)

// DBAssertionFunc is a function that asserts the state of the database.
type DBAssertionFunc = func(*testing.T, *sql.DB, ...any) bool

// AssertDB is a helper function to assert the state of the database.
// This function is useful when you want to assert the state of the database
// after a transaction is committed.
func AssertRecordCount(
	t *testing.T,
	tx *sqlx.Tx,
	expectedCount int,
	table string,
	filterQuery string,
	params ...any,
) bool {
	t.Helper()

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, filterQuery)
	err := tx.GetContext(context.Background(), &count, query, params...)
	if err != nil {
		t.Fatal(err)
	}
	if count != expectedCount {
		t.Errorf(`Expected record count is %d, but got %d
Table: %s
Filter: %s
Params: %v
`, expectedCount, count, table, filterQuery, params)
		return false
	}
	return true
}

// AssertRecordExists is a helper function to assert the state of the database.
// This function is useful when you want to assert the state of the database
// after a transaction is committed.
func AssertRecordExists(
	t *testing.T,
	tx *sqlx.Tx,
	table string,
	filterQuery string,
	params ...any,
) bool {
	return AssertRecordCount(t, tx, 1, table, filterQuery, params...)
}

// AssertRecordNotExists is a helper function to assert the state of the database.
// This function is useful when you want to assert the state of the database
// after a transaction is committed.
func AssertRecordNotExists(
	t *testing.T,
	tx *sqlx.Tx,
	table string,
	filterQuery string,
	params ...any,
) bool {
	return AssertRecordCount(t, tx, 0, table, filterQuery, params...)
}
