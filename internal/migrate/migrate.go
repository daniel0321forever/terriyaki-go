package migrate

import (
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
)

func MigrateDatabase() error {
	err := database.Db.AutoMigrate(
		&models.User{},
		&models.Grind{},
		&models.Task{},
		&models.ParticipateRecord{},
		&models.Message{},
	)
	if err != nil {
		return err
	}

	return nil
}
