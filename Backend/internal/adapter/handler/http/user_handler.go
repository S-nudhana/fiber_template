package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/markbates/goth/gothic"

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
	cookie := c.Cookies("token")
	if cookie == "" {
		return fiber.ErrUnauthorized
	}

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.ErrUnauthorized
	}

	uid, ok := claims["uid"].(string)
	if !ok {
		return fiber.ErrUnauthorized
	}

	c.Locals("uid", uid)
	return c.Next()
}

func (h *HttpUserHandler) BeginOAuth(c *fiber.Ctx) error {
	provider := c.Params("provider")

	return adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(w, r)
	})(c)
}

func (h *HttpUserHandler) OAuthCallback(c *fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, "OAuth failed", http.StatusUnauthorized)
			return
		}

		status, uid, err := h.service.OAuthLogin(
			r.Context(),
			user.Email,
			user.Provider,
			user.FirstName,
			user.LastName,
		)
		if err != nil || !status {
			http.Error(w, "OAuth failed", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{
			"uid": uid,
			"exp": time.Now().Add(72 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			http.Error(w, "Token error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    signed,
			Expires:  time.Now().Add(72 * time.Hour),
			HttpOnly: true,
			Secure:   os.Getenv("ENV") == "production",
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, os.Getenv("ORIGIN"), http.StatusFound)
	})(c)
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
	if err != nil || !status {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    signedToken,
		Expires:  time.Now().Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   os.Getenv("ENV") == "production",
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect request format",
		})
	}

	status, err := h.service.Register(context.Background(), userRegisterPayload.Email, userRegisterPayload.Password, userRegisterPayload.Firstname, userRegisterPayload.Lastname)
	if err != nil || !status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
	})
}

func (h *HttpUserHandler) DeleteUser(c *fiber.Ctx) error {
	uid := c.Locals("uid").(string)
	status, err := h.service.DeleteUser(context.Background(), uid)
	if err != nil || !status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func (h *HttpUserHandler) UpdateUser(c *fiber.Ctx) error {
	uid := c.Locals("uid").(string)
	userUpdateuserPayload := new(domain.UserUpdateRequest)
	if err := c.BodyParser(userUpdateuserPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	if err := validate.Struct(userUpdateuserPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect request format",
		})
	}
	status, err := h.service.UpdateUser(context.Background(), uid, userUpdateuserPayload.Firstname, userUpdateuserPayload.Lastname)
	if err != nil || !status {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Update user data successfully ",
	})
}