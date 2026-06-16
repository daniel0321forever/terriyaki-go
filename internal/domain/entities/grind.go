package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

/** Unit of grind */
type Grind struct {
	ID             string
	Duration       int32 // stored in days
	Participants   []User
	Budget         int32
	Tasks          []HabitTask
	StartDate      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PartnerGroupID string // references PartnerGroup.ID; empty when no group is attached (per D-04)
}

/** Constructor in factory pattern
 * @param duration - the duration of the grind in days
 * @param budget - the budget of the grind in dollars
 * @param participants - the participants of the grind
 * @param startDate - the start date of the grind
 * @return the created grind
 */
func NewGrind(duration int, budget int, startDate time.Time) (*Grind, error) {
	if duration <= 0 {
		return nil, errors.New("duration must be at least 1 day")
	}
	if budget < 0 {
		return nil, errors.New("budget cannot be negative")
	}

	now := time.Now().UTC()

	return &Grind{
		ID:           uuid.New().String(),
		Duration:     int32(duration),
		Participants: []User{},
		Budget:       int32(budget),
		Tasks:        []HabitTask{},
		StartDate:    startDate.UTC(),
		CreatedAt:    now,
		UpdatedAt:    now,
		// Notice: We do NOT initialize Participants or Tasks here
		// if they require further database lookups.
	}, nil
}
