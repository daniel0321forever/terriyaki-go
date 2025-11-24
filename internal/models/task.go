package models

import (
	"encoding/json"
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
	if *task.ProblemTitle != "" {
		return nil
	}

	var problemTitle string
	var problemDescription string
	var problemURL string
	var problemDifficulty string
	var problemTopicTags datatypes.JSON

	startOfTaskDate := task.Date.UTC().Truncate(24 * time.Hour).Add(-time.Hour * 1)
	endOfTaskDate := startOfTaskDate.Add(time.Hour * 24)

	grind, _ := GetGrind(task.GrindID)
	tasks := []Task{}
	database.Db.Where("grind_id = ? AND date >= ? AND date <= ? AND problem_title IS NULL", task.GrindID, startOfTaskDate, endOfTaskDate).Find(&tasks)

	for _, t := range tasks {
		if t.ProblemTitle != nil && *t.ProblemTitle != "" {
			problemTitle = *t.ProblemTitle
			problemDescription = *t.ProblemDescription
			problemURL = *t.ProblemURL
			problemDifficulty = *t.ProblemDifficulty
			problemTopicTags = t.ProblemTopicTags
		}
	}

	// the case that the problem is already set on other participants' tasks
	if problemTitle != "" {
		for _, t := range tasks {
			if t.ProblemTitle != nil && *t.ProblemTitle != "" {
				continue
			}
			t.ProblemTitle = &problemTitle
			t.ProblemDescription = &problemDescription
			t.ProblemURL = &problemURL
			t.ProblemDifficulty = &problemDifficulty
			t.ProblemTopicTags = problemTopicTags

			database.Db.Save(&t)
		}

		return nil
	}

	leetcodeProblem, err := utils.GetRandomLeetCodeProblem()
	if err != nil {
		return err
	}

	problemTitle = leetcodeProblem.Title
	problemDescription = "A daily problem from LeetCode"
	problemURL = "https://leetcode.com/problems/" + leetcodeProblem.TitleSlug + "/description"
	problemDifficulty = leetcodeProblem.Difficulty

	// get topic tag names
	topicTagNames := []string{}
	for _, tag := range leetcodeProblem.TopicTags {
		topicTagNames = append(topicTagNames, tag.Name)
	}
	jsonBytes, err := json.Marshal(topicTagNames)
	if err != nil {
		return err
	}
	problemTopicTags = datatypes.JSON(jsonBytes)

	for _, participant := range grind.Participants {

		var t Task
		database.Db.Where("user_id = ? AND grind_id = ? AND date >= ? AND date <= ?", participant.ID, grind.ID, startOfTaskDate, endOfTaskDate).First(&t)

		t.ProblemTitle = &problemTitle
		t.ProblemDescription = &problemDescription
		t.ProblemURL = &problemURL
		t.ProblemDifficulty = &problemDifficulty
		t.ProblemTopicTags = problemTopicTags

		database.Db.Save(&t)
	}

	return nil
}
