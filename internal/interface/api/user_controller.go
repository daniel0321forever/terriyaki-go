package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	grindService *services.GrindService
	userService  *services.UserService
}

func NewUserController(
	gs *services.GrindService,
	us *services.UserService,
) *UserController {
	return &UserController{
		grindService: gs,
		userService:  us,
	}
}

func (ctrl *UserController) RegisterAPI(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	username := body["username"]
	email := body["email"]
	password := body["password"]
	avatar := body["avatar"]

	createUserDTO := dto.CreateUserDTO{
		Username: username,
		Email:    email,
		Password: password,
		Avatar:   avatar,
	}
	userDTO, err := ctrl.userService.CreateUser(createUserDTO)
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

	token, err := utils.GenerateJWTToken(userDTO.ID)
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
		"user":    userDTO,
		"token":   token,
		"grind":   nil,
	})
}

func (ctrl *UserController) LoginAPI(c *gin.Context) {
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
	getUserDTO := dto.GetUserByEmailDTO{
		Email: email,
	}
	userDTO, err := ctrl.userService.GetUserByEmail(getUserDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid email",
			"errorCode": config.ERROR_CODE_INVALID_EMAIL,
		})
		return
	}

	if !utils.VerifyPassword(password, userDTO.HashedPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid password",
			"errorCode": config.ERROR_CODE_INVALID_PASSWORD,
		})
		return
	}

	token, err := utils.GenerateJWTToken(userDTO.ID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userDTO.ID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    userDTO,
		"token":   token,
		"grind":   grindDTO,
	})
}

func (ctrl *UserController) VerifyTokenAPI(c *gin.Context) {
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

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	userDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}
	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userDTO.ID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified",
		"user":    userDTO,
		"grind":   grindDTO,
	})
}

func (ctrl *UserController) LogoutAPI(c *gin.Context) {
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

// TODO: DeleteUser
