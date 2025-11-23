package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ID                 string         `json:"id" gorm:"primaryKey"`
	TaskType           string         `json:"task_type" gorm:"not null"`
	UserID             string         `json:"user_id" gorm:"not null"`
	GrindID            string         `json:"grind_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	Date               time.Time      `json:"date" gorm:"not null"`
	FinishedTime       time.Time      `json:"finished_time" gorm:""`
	Completed          bool           `json:"completed" gorm:"not null"`
	ProblemTitle       *string        `json:"problem_title" gorm:""`
	ProblemDescription *string        `json:"problem_description" gorm:""`
	ProblemURL         *string        `json:"problem_url" gorm:""`
	ProblemDifficulty  *string        `json:"problem_difficulty" gorm:""`
	ProblemTopicTags   datatypes.JSON `json:"problem_topic_tags" gorm:""`

	Code         *string `json:"code" gorm:""`
	CodeLanguage *string `json:"code_language" gorm:""`
}

/**
 * Create a new task
 * @param title - the title of the task
 * @param description - the description of the task
 * @param url - the url of the task
 * @param progressID - the id of the progress
 * @param date - the date of the task
 * @return the created task
 */
func CreateTask(
	title string,
	description string,
	url string,
	userID string,
	grindID string,
	date time.Time,
) (*Task, error) {
	task := Task{
		ID:                 uuid.New().String(),
		TaskType:           "leetcode",
		ProblemTitle:       &title,
		ProblemDescription: &description,
		ProblemURL:         &url,
		UserID:             userID,
		GrindID:            grindID,
		Date:               date,
		Completed:          false,
	}

	result := database.Db.Create(&task)
	if result.Error != nil {
		return nil, result.Error
	}

	return &task, nil
}

/**
 * Get today task by user id and grind id
 * @param userID - the id of the user
 * @param grindID - the id of the grind
 * @return the today task
 */
func GetTodayTask(userID string, grindID string) (*Task, error) {

	startOfToday := time.Now().UTC().Truncate(24 * time.Hour).Add(-time.Hour * 1)
	endOfToday := startOfToday.Add(time.Hour * 24)

	var task Task
	result := database.Db.Where("user_id = ? AND grind_id = ? AND date >= ? AND date <= ?", userID, grindID, startOfToday, endOfToday).First(&task)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := setTaskProblemIfNeeded(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

/**
 * Get a task by id
 * @param id - the id of the task
 * @return the task
 */
func GetTaskByID(id string, setProblemIfNeeded bool) (*Task, error) {
	var task Task
	result := database.Db.Where("id = ?", id).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}

	if setProblemIfNeeded {
		err := setTaskProblemIfNeeded(&task)
		if err != nil {
			return nil, err
		}
	}

	return &task, nil
}

/**
 * Delete a task
 * @param id - the id of the task
 * @return the result of the deletion
 */
func DeleteTask(id string) error {
	result := database.Db.Unscoped().Delete(&Task{}, id)
	return result.Error
}

func setTaskProblemIfNeeded(task *Task) error {
	if *task.ProblemURL == "" {
		leetcodeProblem, err := utils.GetRandomLeetCodeProblem()
		if err != nil {
			return err
		}

		problemTitle := leetcodeProblem.Title
		problemDescription := "A daily problem from LeetCode"
		problemURL := "https://leetcode.com/problems/" + leetcodeProblem.TitleSlug + "/description"

		fmt.Println("task.GrindID", task.GrindID)
		grind, err := GetGrind(task.GrindID)
		if err != nil {
			return err
		}

		startOfTaskDate := task.Date.UTC().Truncate(24 * time.Hour).Add(-time.Hour * 1)
		endOfTaskDate := startOfTaskDate.Add(time.Hour * 24)

		for _, participant := range grind.Participants {

			var task Task
			database.Db.Where("user_id = ? AND grind_id = ? AND date >= ? AND date <= ?", participant.ID, grind.ID, startOfTaskDate, endOfTaskDate).First(&task)

			task.ProblemTitle = &problemTitle
			task.ProblemDescription = &problemDescription
			task.ProblemURL = &problemURL
			task.ProblemDifficulty = &leetcodeProblem.Difficulty
			topicTagNames := []string{}

			for _, tag := range leetcodeProblem.TopicTags {
				topicTagNames = append(topicTagNames, tag.Name)
			}
			// Marshal topicTagNames to JSON before assigning to ProblemTopicTags
			jsonBytes, err := json.Marshal(topicTagNames)
			if err != nil {
				return err
			}
			task.ProblemTopicTags = datatypes.JSON(jsonBytes)

			database.Db.Save(&task)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
