package models

import (
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"gorm.io/gorm"
)

type ParticipateRecord struct {
	gorm.Model
	UserID       string    `json:"user_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	GrindID      string    `json:"grind_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	MissedDays   int       `json:"missed_days" gorm:"not null;default:0"`
	TotalPenalty int       `json:"total_penalty" gorm:"not null;default:0"`
	Quitted      bool      `json:"quitted" gorm:"not null;default:false"`
	QuittedAt    time.Time `json:"quitted_at" gorm:""`
}

func (ParticipateRecord) TableName() string {
	return "participate_records"
}

/**
 * Create a new participant
 * @param userID - the id of the user
 * @param grindID - the id of the grind
 * @return the created participant record
 */
func CreateParticipateRecord(userID string, grindID string) (*ParticipateRecord, error) {
	participateRecord := ParticipateRecord{
		UserID:       userID,
		GrindID:      grindID,
		MissedDays:   0,
		TotalPenalty: 0,
		Quitted:      false,
		QuittedAt:    time.Time{},
	}

	result := database.Db.FirstOrCreate(&participateRecord, ParticipateRecord{UserID: userID, GrindID: grindID})
	if result.Error != nil {
		return nil, result.Error
	}

	return &participateRecord, nil
}

/**
 * Get a participant record by id
 * @param id - the id of the participant record
 * @return the participant record
 */
func GetParticipateRecord(id string) (*ParticipateRecord, error) {
	var participateRecord ParticipateRecord
	result := database.Db.Where("id = ?", id).First(&participateRecord)
	if result.Error != nil {
		return nil, result.Error
	}
	return &participateRecord, nil
}

/**
 * Get a participant record by user id and grind id
 * @param userID - the id of the user
 * @param grindID - the id of the grind
 * @return the participant record, `userID`, `grindID`, `missedDays`, `totalPenalty`, `quitted` and `quittedAt`
 */
func GetParticipateRecordByUserIDAndGrindID(userID string, grindID string) (*ParticipateRecord, error) {
	// find all participant record
	fmt.Println("GetParticipantRecordByUserIDAndGrindID", userID, grindID)
	var allRecords []ParticipateRecord
	database.Db.Find(&allRecords)
	for _, record := range allRecords {
		fmt.Println(record.UserID, record.GrindID)
	}

	var participateRecord ParticipateRecord
	result := database.Db.Where("user_id = ? AND grind_id = ?", userID, grindID).First(&participateRecord)
	if result.Error != nil {
		return nil, result.Error
	}
	return &participateRecord, nil
}

/**
 * Update a participant record
 * @param id - the id of the participant record
 * @param missedDays - the number of missed days
 * @param totalPenalty - the total penalty
 * @param quitted - whether the participant has quitted
 * @param quittedAt - the date and time the participant quitted
 * @return the updated participant record
 */
func UpdateParticipateRecord(
	id string,
	updates map[string]any,
) (*ParticipateRecord, error) {
	var participateRecord ParticipateRecord
	result := database.Db.Model(&participateRecord).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	return &participateRecord, nil
}
