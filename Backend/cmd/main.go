package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	adapter "github.com/S-nudhana/fiber_template/internal/adapter/database"
	httpHandler "github.com/S-nudhana/fiber_template/internal/adapter/handler/http"
	"github.com/S-nudhana/fiber_template/internal/adapter/handler/router"
	"github.com/S-nudhana/fiber_template/internal/core/service"
	"github.com/S-nudhana/fiber_template/internal/infrastructure/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		Secure:   false,                // MUST be false on localhost
		SameSite: http.SameSiteLaxMode, // REQUIRED
	}
	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:3000/api/user/oauth/google/callback",
		),
	)
	db, err := database.New()
	if err != nil {
		log.Fatal("failed connecting to db:", err)
	}
	defer db.Close()

	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "${ip}:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ORIGIN"),
		AllowCredentials: true,
	}))

	userRepo := adapter.NewMySQLUserAdapter(db)
	userService := service.NewUserService(userRepo)
	userHandler := httpHandler.NewHttpUserHandler(userService)

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "API is working!"})
	})

	router.UserRouter(app, userHandler)

	addr := ":3000"
	log.Printf("Server running at http://localhost%s\n", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatal("Server stopped:", err)
	}
}
