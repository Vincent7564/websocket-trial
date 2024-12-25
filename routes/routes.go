package routes

import (
	"websocket-trial/controller"
	"websocket-trial/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

type Router struct {
	App *fiber.App
	DB  *gorm.DB
}

func (r *Router) Init() {
	controller := controller.Controller{DB: r.DB}
	AuthRoutes(r.App, controller)
}

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	handler := handler.Handler{DB: db}
	app.Get("/ws", websocket.New(handler.EchoServer))

	router := Router{App: app, DB: db}
	router.Init()
}
