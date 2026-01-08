package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"not null"`
	Email     string    `json:"email" gorm:"not null;unique"`
	Avatar    string    `json:"avatar" gorm:""`
	Password  string    `json:"password" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

/**
 * Create a new user, hash the password and create a new user record in the database
 * @param username - the username of the user
 * @param email - the email of the user
 * @param password - the password of the user
 * @param avatar - the avatar of the user
 * @return the created user
 */
func CreateUser(
	username string,
	email string,
	password string,
	avatar string,
) (*User, error) {
	fmt.Println("Creating user: ", username, email, password, avatar)

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Avatar:    avatar,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := database.Db.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

/**
 * Get a user by id
 * @param id - the id of the user
 * @return the user
 */
func GetUser(id string) (*User, error) {
	var user User
	result := database.Db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "record not found") {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

/**
 * Get a user by email
 * @param email - the email of the user
 * @return the user
 */
func GetUserByEmail(email string) (*User, error) {
	var user User
	result := database.Db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

/**
 * Update a user, only update the fields that are provided
 * @param id - the id of the user
 * @param username - the username of the user
 * @param email - the email of the user
 * @param avatar - the avatar of the user
 * @param password - the password of the user
 * @return the updated user
 */
func UpdateUser(
	id string,
	username *string,
	email *string,
	avatar *string,
	password *string,
) (*User, error) {
	var user User
	updates := make(map[string]any)
	if username != nil && *username != "" && *username != "none" {
		updates["username"] = *username
	}
	if email != nil && *email != "" && *email != "none" {
		updates["email"] = *email
	}
	if avatar != nil && *avatar != "" && *avatar != "none" {
		updates["avatar"] = *avatar
	}
	if password != nil && *password != "" && *password != "none" {
		updates["password"] = *password
	}
	result := database.Db.Model(&user).Where("id = ?", id).Updates(updates)
	return &user, result.Error
}

/**
 * Delete a user
 * @param id - the id of the user
 * @return the result of the deletion
 */
func DeleteUser(id string) error {
	// Find the user first
	var user User
	if err := database.Db.Where("id = ?", id).First(&user).Error; err != nil {
		return err
	}

	// Delete all tasks associated with this user
	if err := database.Db.Where("user_id = ?", id).Unscoped().Delete(&Task{}).Error; err != nil {
		return err
	}

	// Remove user from all grind_participants associations (many-to-many join table)
	// Since the relationship is defined on Grind side, we delete directly from the join table
	if err := database.Db.Exec("DELETE FROM participate_records WHERE user_id = ?", id).Error; err != nil {
		return err
	}

	// Now delete the user
	result := database.Db.Unscoped().Delete(&user)
	return result.Error
}
