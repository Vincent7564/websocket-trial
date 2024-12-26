package main

import (
	"fmt"
	"log"
	"websocket-trial/models"
	"websocket-trial/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Go Websocket Server")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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
		&models.UserAccessToken{},
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
	app.Use(cors.New())
	routes.SetupRoutes(app, db)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
