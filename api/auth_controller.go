package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAPI(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	username := body["username"]
	email := body["email"]
	password := body["password"]
	avatar := body["avatar"]

	user, err := models.CreateUser(username, email, password, avatar)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			c.JSON(http.StatusBadRequest, gin.H{
				"message":   "email already exists",
				"errorCode": config.ERROR_CODE_DUPLICATE_ENTRY,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
		"user":    serializer.SerializeUser(user),
		"token":   token,
		"grind":   nil,
	})
}

func LoginAPI(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid request body",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	email := body["email"]
	password := body["password"]

	// get user by email
	user, err := models.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid email",
			"errorCode": config.ERROR_CODE_INVALID_EMAIL,
		})
		return
	}

	if !utils.VerifyPassword(password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid password",
			"errorCode": config.ERROR_CODE_INVALID_PASSWORD,
		})
		return
	}

	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	grind, err := models.GetOngoingGrindByUserID(user.ID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"message":   "No current grind found",
				"user":      serializer.SerializeUser(user),
				"token":     token,
				"grind":     nil,
				"errorCode": config.ERROR_CODE_NOT_FOUND,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    serializer.SerializeUser(user),
		"token":   token,
		"grind":   serializer.SerializeGrind(user, grind, false),
	})
}

func LoginAPIV2(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid request body",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	email := body["email"]
	password := body["password"]

	// get user by email
	user, err := models.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid email",
			"errorCode": config.ERROR_CODE_INVALID_EMAIL,
		})
		return
	}

	if !utils.VerifyPassword(password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid password",
			"errorCode": config.ERROR_CODE_INVALID_PASSWORD,
		})
		return
	}

	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	grinds, err := models.GetAllUserGrinds(user.ID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"message":   "No current grind found",
				"user":      serializer.SerializeUser(user),
				"token":     token,
				"grinds":    make(map[string]gin.H),
				"errorCode": config.ERROR_CODE_NOT_FOUND,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    serializer.SerializeUser(user),
		"token":   token,
		"grinds":  serializer.SerializeGrindsInMap(user, grinds, false),
	})
}

func VerifyTokenAPI(c *gin.Context) {
	fmt.Println("VerifyTokenAPI")
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "No user found for the token",
				"user":    nil,
				"grind":   nil,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	grind, err := models.GetOngoingGrindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Token verified with no current grind found",
				"user":    serializer.SerializeUser(user),
				"grind":   nil,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified",
		"user":    serializer.SerializeUser(user),
		"grind":   serializer.SerializeGrind(user, grind, false),
	})
}

func VerifyTokenAPIV2(c *gin.Context) {
	fmt.Println("VerifyTokenAPI")
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "No user found for the token",
				"user":    nil,
				"grinds":  make(map[string]gin.H),
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	grinds, err := models.GetAllUserGrinds(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Token verified with no grinds found",
				"user":    serializer.SerializeUser(user),
				"grinds":  make(map[string]gin.H),
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified",
		"user":    serializer.SerializeUser(user),
		"grinds":  serializer.SerializeGrindsInMap(user, grinds, false),
	})
}

func LogoutAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func DeleteUserAPI(c *gin.Context) {
	// TODO: remove it after testing
	var users []models.User
	result := database.Db.Find(&users)
	if result.Error != nil {
		fmt.Println("Error fetching users:", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	var deleted []string
	var notFound []string

	for _, user := range users {
		err := models.DeleteUser(user.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Println("User not found:", user.ID)
				notFound = append(notFound, user.ID)
				continue
			}
			fmt.Println("Error deleting user:", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "internal server error",
				"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
			})
			return
		}
		deleted = append(deleted, user.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "All users processed",
		"deleted":   deleted,
		"not_found": notFound,
	})
}

// GET /api/v1/users/exists?email=xxx@example.com
func CheckUserExistsAPI(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "missing email query parameter",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	_, err := models.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "User does not exist",
			"exists":  false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User exists",
		"exists":  true,
	})
}
