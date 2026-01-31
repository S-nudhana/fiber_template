package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/S-nudhana/fiber_template/internal/adapter/handler/http"
)

func UserRouter(app *fiber.App, userHandler *http.HttpUserHandler) {
	user := app.Group("/api/user")

	user.Post("/login", userHandler.Login)
	user.Post("/register", userHandler.Register)
	user.Get("/oauth/:provider", userHandler.BeginOAuth)
	user.Get("/oauth/:provider/callback", userHandler.OAuthCallback)

	user.Use(http.AuthRequired)
	user.Delete("/delete", userHandler.DeleteUser)
	user.Put("/update", userHandler.UpdateUser)
}