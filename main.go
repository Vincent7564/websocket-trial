package main

import (
	"fmt"
	"log"
	"websocket-trial/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Go Websocket Server")

	app := fiber.New()

	routes.SetupRoutes(app) // Setup routes using the routes package

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
