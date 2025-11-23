package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Grind struct {
	gorm.Model
	ID           string    `json:"id" gorm:"primaryKey"`
	Duration     int32     `json:"duration" gorm:"not null"` // stored in days
	Participants []User    `json:"participants" gorm:"many2many:participate_records;foreignKey:ID;references:ID"`
	Budget       int32     `json:"budget" gorm:"not null"`
	Tasks        []Task    `json:"tasks" gorm:"not null"`
	StartDate    time.Time `json:"start_date" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null"`
}

/**
 * Create a new grind
 * @param duration - the duration of the grind in days
 * @param budget - the budget of the grind in dollars
 * @param participants - the participants of the grind
 * @param startDate - the start date of the grind
 * @return the created grind
 */
func CreateGrind(
	duration int,
	budget int,
	participants []any,
	startDate time.Time,
) (*Grind, error) {
	participantsUsers := []User{}
	for _, participant := range participants {
		user, err := GetUserByEmail(participant.(string))
		if err != nil {
			return nil, errors.New(config.ERROR_CODE_USER_NOT_FOUND)
		}
		participantsUsers = append(participantsUsers, *user)
	}

	grind := Grind{
		ID:           uuid.New().String(),
		Duration:     int32(duration),
		Budget:       int32(budget),
		Participants: participantsUsers,
		StartDate:    startDate,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result := database.Db.Create(&grind)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, user := range participantsUsers {
		_, err := CreateParticipateRecord(user.ID, grind.ID)
		if err != nil {
			return nil, err
		}
		for i := range duration {
			_, err := CreateTask(
				"Task "+strconv.Itoa(i+1),
				"Task "+strconv.Itoa(i+1)+" description",
				"",
				user.ID,
				grind.ID,
				startDate.AddDate(0, 0, i),
			)

			if err != nil {
				return nil, err
			}
		}
	}

	return &grind, nil
}

/**
 * Get a grind by id
 * @param id - the id of the grind
 * @return the grind
 */
func GetGrind(id string) (*Grind, error) {
	var grind Grind
	result := database.Db.Preload("Participants").Where("id = ?", id).First(&grind)
	if result.Error != nil {
		return nil, result.Error
	}
	return &grind, nil
}

/**
 * Get a grind by user id, only return the ginrd with end date after current date
 * @param userID - the id of the user
 * @return the grind
 */
func GetOngoingGrindByUserID(userID string) (*Grind, error) {
	var grind Grind
	result := database.Db.Preload("Participants").
		Joins("JOIN participate_records ON participate_records.grind_id = grinds.id").
		Where("participate_records.user_id = ?", userID).
		Order("grinds.created_at DESC").
		First(&grind)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Check if the grind has ended
	endDate := grind.StartDate.AddDate(0, 0, int(grind.Duration))
	if endDate.Before(time.Now().UTC()) {
		return nil, gorm.ErrRecordNotFound
	}

	// Check the participate record for the current user
	var participateRecord ParticipateRecord
	result = database.Db.Where("user_id = ? AND grind_id = ?", userID, grind.ID).First(&participateRecord)
	if result.Error != nil {
		// If the participate record is not found, return not found error
		return nil, result.Error
	}

	if participateRecord.Quitted {
		return nil, gorm.ErrRecordNotFound
	}

	return &grind, nil
}

/**
 * Update a grind
 * @param id - the id of the grind
 * @param duration - the duration of the grind in days
 * @param budget - the budget of the grind in dollars
 * @return the updated grind
 */
func UpdateGrind(id string, updates map[string]any) (*Grind, error) {

	var grind Grind
	result := database.Db.Preload("Participants").Model(&grind).Where("id = ?", id).First(&grind).Updates(updates)

	if result.Error != nil {
		return nil, result.Error
	}
	return &grind, nil
}

func AddParticipantToGrind(grindID string, participantID string) error {
	participateRecord, _ := GetParticipateRecordByUserIDAndGrindID(participantID, grindID)
	if participateRecord != nil {
		return errors.New("PARTICIPANT_EXISTS")
	}

	_, err := CreateParticipateRecord(participantID, grindID)
	if err != nil {
		return err
	}

	return nil
}

func RemoveParticipantFromGrind(grindID string, participantID string) error {
	participateRecord, _ := GetParticipateRecordByUserIDAndGrindID(participantID, grindID)
	if participateRecord == nil {
		return errors.New("PARTICIPANT_NOT_FOUND")
	}

	result := database.Db.Unscoped().Delete(&participateRecord)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func GetAllUserGrinds(userID string) ([]Grind, error) {
	var grinds []Grind
	result := database.Db.Preload("Participants").
		Joins("JOIN participate_records ON participate_records.grind_id = grinds.id").
		Where("participate_records.user_id = ?", userID).
		Find(&grinds)
	if result.Error != nil {
		return nil, result.Error
	}
	return grinds, nil
}

func GetAllGrinds() ([]Grind, error) {
	var grinds []Grind
	result := database.Db.Preload("Participants").Find(&grinds)
	if result.Error != nil {
		return nil, result.Error
	}
	return grinds, nil
}

/**
 * Delete a grind
 * @param id - the id of the grind
 * @return the result of the deletion
 */
func DeleteGrind(id string) error {
	var grind Grind
	result := database.Db.Where("id = ?", id).First(&grind)
	if result.Error != nil {
		return result.Error
	}

	// Delete all associated tasks first
	result = database.Db.Where("grind_id = ?", id).Unscoped().Delete(&Task{})
	if result.Error != nil {
		return result.Error
	}

	// Clear the many-to-many associations
	err := database.Db.Model(&grind).Association("Participants").Clear()
	if err != nil {
		return err
	}

	result = database.Db.Unscoped().Delete(&grind)
	return result.Error
}

/**
 * Delete all grinds
 * @return the result of the deletion
 */
func DeleteAllGrinds() error {
	var grinds []Grind
	result := database.Db.Find(&grinds)
	if result.Error != nil {
		return result.Error
	}

	for _, grind := range grinds {
		err := DeleteGrind(grind.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
