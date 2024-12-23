package routes

import (
	"websocket-trial/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	handler := handler.Handler{DB: db}
	app.Get("/ws", websocket.New(handler.EchoServer))
}
