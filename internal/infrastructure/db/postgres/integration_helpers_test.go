//go:build integration
// +build integration

package postgres

import (
	"testing"
)

func resetRepoTables(t *testing.T) {
	t.Helper()
	err := Db.Exec("TRUNCATE TABLE participation, grinds, users RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("failed to reset test tables: %v", err)
	}
}
