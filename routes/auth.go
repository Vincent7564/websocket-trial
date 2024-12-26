package routes

import (
	"websocket-trial/controller"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App, controller controller.Controller) {
	AuthRoutes := app.Group("/auth")

	AuthRoutes.Post("/register", controller.Register)
	AuthRoutes.Post("/login", controller.Login)
	// AuthRoutes.Post("/forgot-password", controller.ForgotPassword)
	// AuthRoutes.Post("/check-forgot-token", controller.CheckForgetPasswordToken)
	// AuthRoutes.Post("/reset-password", controller.ResetPassword)
}
