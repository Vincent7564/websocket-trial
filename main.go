package main

import (
	"fmt"
	"log"
	"websocket-trial/models"
	"websocket-trial/routes"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Go Websocket Server")

	conn := "host=localhost user=postgres dbname=RTChat sslmode=disable"
	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to Database")
	}

	models := []interface{}{
		&models.Chat{},
		&models.Channel{},
		&models.ChannelMember{},
		&models.User{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate model: %v", err)
		}
	}

	defer func() {
		dbInstances, _ := db.DB()
		_ = dbInstances.Close()
	}()

	app := fiber.New()
	routes.SetupRoutes(app, db)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
