package routes

import (
	"websocket-trial/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/ws", websocket.New(handler.EchoServer))
}
