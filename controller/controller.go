package controller

import (
	"net/http"
	"websocket-trial/models"
	"websocket-trial/models/dto"
	"websocket-trial/utils"

	"github.com/gofiber/fiber/v2"
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
