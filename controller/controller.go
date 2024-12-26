package controller

import (
	"net/http"
	"os"
	"time"
	"websocket-trial/models"
	"websocket-trial/models/dto"
	"websocket-trial/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Controller struct {
	DB *gorm.DB
}

func (c *Controller) Register(ctx *fiber.Ctx) error {
	var request dto.RegisterRequest
	err := ctx.BodyParser(&request)

	if err != nil {
		return utils.GenerateResponse(ctx, http.StatusBadGateway, "Invalid Request", err.Error())
	}

	if errorMessage := utils.ValidateData(&request); len(errorMessage) > 0 {
		return utils.GenerateResponse(ctx, http.StatusBadGateway, "Validation Error, Please More Carefully Insert The Data", errorMessage)
	}

	hash, _ := utils.HashPassword(request.Password)
	request.Password = hash
	user := models.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	}
	if err := c.DB.Create(&user).Error; err != nil {
		return utils.GenerateResponse(ctx, http.StatusInternalServerError, "Failed to insert data", err.Error())
	}
	return utils.GenerateResponse(ctx, http.StatusOK, "Register Success", &user)
}

func (c *Controller) Login(ctx *fiber.Ctx) error {
	var request dto.LoginRequest
	err := ctx.BodyParser(&request)
	var dataRetrieved models.User
	if err != nil {
		return utils.GenerateResponse(ctx, http.StatusBadGateway, "Invalid Request", err.Error())
	}

	if errorMessage := utils.ValidateData(&request); len(errorMessage) > 0 {
		return utils.GenerateResponse(ctx, http.StatusBadGateway, "Validation Error, Please More Carefully Insert The Data", errorMessage)
	}

	if err := c.DB.Where("username = ?", request.Username).First(&dataRetrieved).Error; err != nil {
		return utils.GenerateResponse(ctx, http.StatusInternalServerError, "Failed to find data", err.Error())
	}

	isTrue := utils.CheckPasswordHash(request.Password, dataRetrieved.Password)

	if isTrue {
		claims := jwt.MapClaims{
			"username": dataRetrieved.Username,
			"email":    dataRetrieved.Email,
			"exp":      time.Now().Add(time.Hour * 3).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		secret_token := os.Getenv("JWT_SECRET_KEY")
		t, err := token.SignedString([]byte(secret_token))

		if err != nil {
			return utils.GenerateResponse(ctx, http.StatusInternalServerError, "Failed to Sign Token", err)
		}

		accessToken := models.UserAccessToken{
			UserID:    dataRetrieved.ID,
			Token:     t,
			CreatedAt: time.Now(),
			ExpiredAt: time.Now().Add(time.Hour * 3),
		}

		if err := c.DB.Create(&accessToken).Error; err != nil {
			return utils.GenerateResponse(ctx, http.StatusInternalServerError, "Failed to insert token", err.Error())
		}
		return utils.GenerateResponse(ctx, http.StatusOK, "Generate Token Success", &accessToken)
	}
	return utils.GenerateResponse(ctx, http.StatusBadGateway, "Failed Generate Token", "")
}
