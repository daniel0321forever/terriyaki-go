//go:build integration
// +build integration

package postgres_test

import (
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
)

func resetRepoTables(t *testing.T) {
	t.Helper()
	err := postgres.Db.Exec("TRUNCATE TABLE participation, grinds, users RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("failed to reset test tables: %v", err)
	}
}
