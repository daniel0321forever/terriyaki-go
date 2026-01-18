package migrate

import (
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
)

func MigrateDatabase() error {
	err := postgres.Db.AutoMigrate(
		&postgres.UserSchema{},
		&postgres.GrindSchema{},
		&postgres.TaskSchema{},
		&postgres.ParticipationSchema{},
		&postgres.MessageSchema{},
		&postgres.InterviewSessionSchema{},
	)
	if err != nil {
		return err
	}

	return nil
}
