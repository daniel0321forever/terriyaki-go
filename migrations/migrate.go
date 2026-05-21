package migrations

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	mg "github.com/golang-migrate/migrate/v4"
	pgdriver "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

const (
	sqlMigrationsDirName = "sql"
	fileSourcePrefix     = "file://"
)

func MigrateDatabase(db *gorm.DB) error {
	return MigrateUp(db)
}

func MigrateUp(db *gorm.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil && !errors.Is(err, mg.ErrNoChange) {
		return err
	}

	// schema is up to date
	return nil
}

func MigrateDown(db *gorm.DB, steps int) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if steps > 0 {
		if err := m.Steps(-steps); err != nil && !errors.Is(err, mg.ErrNoChange) {
			return err
		}
		return nil
	}

	if err := m.Down(); err != nil && !errors.Is(err, mg.ErrNoChange) {
		return err
	}

	return nil
}

func newMigrator(db *gorm.DB) (*mg.Migrate, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	driver, err := pgdriver.WithInstance(sqlDB, &pgdriver.Config{})
	if err != nil {
		return nil, err
	}

	sqlMigrationsDir, err := resolveSQLMigrationsDir()
	if err != nil {
		return nil, err
	}

	sqlMigrationsSourceURL := fmt.Sprintf("%s%s", fileSourcePrefix, filepath.ToSlash(sqlMigrationsDir))
	return mg.NewWithDatabaseInstance(sqlMigrationsSourceURL, "postgres", driver)
}

func closeMigrator(m *mg.Migrate) {
	_ = m
}

func resolveSQLMigrationsDir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to resolve SQL migrations directory")
	}

	sqlMigrationPath := filepath.Join(filepath.Dir(currentFile), sqlMigrationsDirName)
	return sqlMigrationPath, nil
}
