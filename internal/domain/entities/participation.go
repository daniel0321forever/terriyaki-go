package entities

import (
	"time"

	"github.com/google/uuid"
)

/** Progress status (the relationship between User and Grind) */
type Participation struct {
	ID           string
	UserID       string
	GrindID      string
	MissedDays   int
	TotalPenalty int
	Quitted      bool
	QuittedAt    time.Time
}

/** Constructor in factory pattern
 * @param userID - user id
 * @param grindID - grind id
 */
func NewParticipation(userID string, grindID string) (*Participation, error) {
	return &Participation{
		ID:           uuid.New().String(),
		UserID:       userID,
		GrindID:      grindID,
		MissedDays:   0,
		TotalPenalty: 0,
		Quitted:      false,
		QuittedAt:    time.Time{},
	}, nil
}
