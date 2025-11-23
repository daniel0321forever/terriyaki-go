package migrate

import (
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
)

func MigrateDatabase() error {
	err := database.Db.AutoMigrate(&models.User{}, &models.Grind{}, &models.Task{}, &models.ParticipateRecord{})
	if err != nil {
		return err
	}

	// TODO: can remove this after first migration is complete
	err = MigrateParticipateRecord()
	if err != nil {
		return err
	}

	return nil
}
