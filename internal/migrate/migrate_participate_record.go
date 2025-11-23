package migrate

import (
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
)

func MigrateParticipateRecord() error {
	var count int64
	result := database.Db.Table("participate_records").Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count > 0 {
		fmt.Println("Participate records already exist, skipping MigrateParticipateRecord")
		return nil
	}

	// Find all records from old grind_participants table and migrate to participate_records
	type OldGrindParticipant struct {
		UserID  string
		GrindID string
	}

	var oldParticipants []OldGrindParticipant
	result = database.Db.Table("grind_participants").Find(&oldParticipants)
	if result.Error != nil {
		return result.Error
	}

	for _, old := range oldParticipants {
		_, err := models.CreateParticipateRecord(old.UserID, old.GrindID)
		if err != nil {
			fmt.Println("Error creating participant record: ", err)
			return err
		}
	}

	fmt.Println("Migrated participate records successfully")

	return nil
}
