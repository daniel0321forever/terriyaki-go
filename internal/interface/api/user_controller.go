package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		RespondBadRequest(c, "invalid request body")
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
			RespondConflict(c, "email already exists")
			return
		}

		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	token, err := utils.GenerateJWTToken(userDTO.ID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
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
		RespondBadRequest(c, "invalid request body")
		return
	}

	email := body["email"]
	password := body["password"]

	// get user by email
	getUserDTO := dto.GetUserByEmailDTO{
		Email: email,
	}

	// get user
	userDTO, err := ctrl.userService.GetUserByEmail(getUserDTO)
	if err != nil {
		RespondBadRequest(c, "invalid email")
		return
	}

	// verify password
	if !utils.VerifyPassword(password, userDTO.HashedPassword) {
		RespondBadRequest(c, "invalid password")
		return
	}

	// generate token
	token, err := utils.GenerateJWTToken(userDTO.ID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	// get ongoing grind
	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userDTO.ID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)

	if err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    userDTO,
		"token":   token,
		"grind":   grindDTO,
	})
}

func (ctrl *UserController) LoginAPIV2(c *gin.Context) {
	Request := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := c.ShouldBindJSON(&Request); err != nil {
		RespondBadRequest(c, "invalid request body")
		return
	}

	email := Request.Email
	password := Request.Password

	// get user by email
	getUserDTO := dto.GetUserByEmailDTO{
		Email: email,
	}

	// get user
	userDTO, err := ctrl.userService.GetUserByEmail(getUserDTO)
	if err != nil {
		RespondBadRequest(c, "invalid email")
		return
	}

	// verify password
	if !utils.VerifyPassword(password, userDTO.HashedPassword) {
		RespondBadRequest(c, "invalid password")
		return
	}

	// generate token
	token, err := utils.GenerateJWTToken(userDTO.ID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	getGrindsDTO := dto.GetAllUserGrindsDTO{
		UserID: userDTO.ID,
	}
	grinds, err := ctrl.grindService.GetAllUserGrinds(getGrindsDTO)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("No current grind found")
			c.JSON(http.StatusOK, gin.H{
				"message": "No current grind found",
				"user":    userDTO,
				"token":   token,
				"grinds":  make(map[string]gin.H),
			})
			return
		}

		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    userDTO,
		"token":   token,
		"grinds":  grinds,
	})
}

func (ctrl *UserController) VerifyTokenAPI(c *gin.Context) {
	fmt.Println("VerifyTokenAPI")
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	userDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userDTO.ID,
	}

	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified",
		"user":    userDTO,
		"grind":   grindDTO,
	})
}

func (ctrl *UserController) VerifyTokenAPIV2(c *gin.Context) {
	fmt.Println("VerifyTokenAPI")
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	userDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	getGrindsDTO := dto.GetAllUserGrindsDTO{
		UserID: userDTO.ID,
	}

	grinds, err := ctrl.grindService.GetAllUserGrinds(getGrindsDTO)

	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "no grinds found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified",
		"user":    userDTO,
		"grinds":  grinds,
	})
}

func (ctrl *UserController) LogoutAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GET /api/v1/users/exists?email=xxx@example.com
func (ctrl *UserController) CheckUserExistsAPI(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		RespondBadRequest(c, "missing email query parameter")
		return
	}

	getUserDTO := dto.GetUserByEmailDTO{
		Email: email,
	}

	_, err := ctrl.userService.GetUserByEmail(getUserDTO)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "User does not exist", "exists": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User exists", "exists": true})
}
