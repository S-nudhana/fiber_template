package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/S-nudhana/fiber_template/internal/infrastructure/database"
	"github.com/S-nudhana/fiber_template/internal/adapter/database"
	"github.com/S-nudhana/fiber_template/internal/adapter/handler/http"
	"github.com/S-nudhana/fiber_template/internal/adapter/handler/router"
	"github.com/S-nudhana/fiber_template/internal/core/service"
)

func main() {
	app := fiber.New()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := database.New()
	if err != nil {
		log.Fatal("failed connecting to db:", err)
	}
	defer db.Close()

	userRepo := adapter.NewMySQLUserAdapter(db)

	app.Use(logger.New(logger.Config{Format: "${ip}:${port} ${status} - ${method} ${path}\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ORIGIN"),
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowCredentials: true,
	}))

	userService := service.NewUserService(userRepo)
	userHandler := http.NewHttpUserHandler(userService)

	app.Get("/api/test", func(c *fiber.Ctx) error { 
		return c.JSON(fiber.Map{"message": "API is working!"}) 
	})
	router.UserRouter(app, userHandler)

	app.Listen(":3000")
}
