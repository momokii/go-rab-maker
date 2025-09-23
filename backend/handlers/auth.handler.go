package handlers

import (
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// views rendedr
func (h *AuthHandler) LoginView(c *fiber.Ctx) error {
	return adaptor.HTTPHandler(templ.Handler(components.LoginPage()))(c)
}

// function
func (h *AuthHandler) Login(c *fiber.Ctx) error {

	time.Sleep(1 * time.Second)

	username := c.FormValue("username")
	password := c.FormValue("password")

	if username != "admin" || password != "admin" {
		return utils.ResponseErrorModal(c, "Login Failed", "Invalid username or password")
	}

	loginError := false
	if loginError {
		return utils.ResponseErrorModal(c, "Login Failed", "Invalid username or password")
	}

	// create session
	sessionData := models.SessionUser{
		ID:    1,
		Email: "admin@admin.com",
		Role:  "admin",
	}

	// create session id
	sessionId, err := utils.GenerateTokenSession()
	if err != nil {
		return utils.ResponseErrorModal(c, "Login Failed", "Failed to create session")
	}

	// better auth process the base session data will be contain 2 data, that id user that login and the session id
	middlewares.CreateSession(c, middlewares.SESSION_USER_ID, sessionData.ID)
	middlewares.CreateSession(c, middlewares.SESSION_ID, sessionId)

	return utils.ResponseRedirectHTMX(c, "/", fiber.StatusAccepted)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	time.Sleep(5 * time.Second)
	middlewares.DeleteSession(c)

	// redirect to login page
	return utils.ResponseRedirectHTMX(c, "/login", fiber.StatusSeeOther)
}
