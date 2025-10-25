package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"context"
	"os"
	"time"

	"github.com/S-nudhana/fiber_template/internal/core/domain"
	"github.com/S-nudhana/fiber_template/internal/core/service"
)

type HttpUserHandler struct {
	service service.UserService
}

func NewHttpUserHandler(service service.UserService) *HttpUserHandler {
	return &HttpUserHandler{service: service}
}

var validate = validator.New()

func AuthRequired(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	claim, status := token.Claims.(jwt.MapClaims)

	if !status {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	c.Locals("uid", claim["uid"])

	return c.Next()
}

func (h *HttpUserHandler) Login(c *fiber.Ctx) error {
	userLoginPayload := new(domain.UserLoginRequest)
	if err := c.BodyParser(userLoginPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	if err := validate.Struct(userLoginPayload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Incorrect request format",
		})
	}

	status, uid, err := h.service.Login(context.Background(), userLoginPayload.Email, userLoginPayload.Password)
	if err != nil || status {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err,
		})
	}

	claims := jwt.MapClaims{
		"uid": uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    signedToken,
		Expires:  time.Now().Add(72 * time.Hour),
		HTTPOnly: os.Getenv("ENV") == "production",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
	})
}

func (h *HttpUserHandler) Register(c *fiber.Ctx) error {
	userRegisterPayload := new(domain.UserRegisterRequest)
	if err := c.BodyParser(userRegisterPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	if err := validate.Struct(userRegisterPayload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Incorrect request format",
		})
	}

	status, err := h.service.Register(context.Background(), userRegisterPayload.Email, userRegisterPayload.Password, userRegisterPayload.Firstname, userRegisterPayload.Lastname)
	if err != nil || status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
	})
}
