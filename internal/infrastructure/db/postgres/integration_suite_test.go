//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/migrations"
	tc "github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testContainer *pgcontainer.PostgresContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := pgcontainer.Run(
		ctx,
		"postgres:16-alpine",
		pgcontainer.WithDatabase("habitat_test"),
		pgcontainer.WithUsername("habitat"),
		pgcontainer.WithPassword("habitat"),
		tc.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		panic(err)
	}

	testContainer = container

	dsn, err := testContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	Db = db

	if err := migrations.MigrateUp(Db); err != nil {
		panic(err)
	}

	exitCode := m.Run()

	_ = testContainer.Terminate(ctx)

	if exitCode != 0 {
		panic("repository integration tests failed")
	}

	os.Exit(exitCode)
}
