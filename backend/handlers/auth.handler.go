package handlers

import (
	"database/sql"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/repository/users"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type AuthHandler struct {
	dbService databases.SQLiteServices
	usersRepo *users.UsersRepo
}

func NewAuthHandler(dbService databases.SQLiteServices) *AuthHandler {
	return &AuthHandler{
		dbService: dbService,
		usersRepo: users.NewUsersRepo(),
	}
}

// views rendedr
func (h *AuthHandler) LoginView(c *fiber.Ctx) error {
	return adaptor.HTTPHandler(templ.Handler(components.LoginPage()))(c)
}

// Login handles user authentication
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	time.Sleep(1 * time.Second) // Simulate processing time for better UX

	username := c.FormValue("username")
	password := c.FormValue("password")

	// Validate input
	if username == "" || password == "" {
		return utils.ResponseErrorModal(c, "Login Failed", "Username and password are required")
	}

	var user models.User
	var err error

	// Find user in database
	if _, err = h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		user, err = h.usersRepo.FindByUsername(tx, username)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusUnauthorized, fiber.NewError(fiber.StatusUnauthorized, "Invalid username or password")
			}
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Login Failed", "Invalid username or password")
	}

	// Verify password
	if err := models.CheckPassword(user.Password, password); err != nil {
		return utils.ResponseErrorModal(c, "Login Failed", "Invalid username or password")
	}

	// Create session data (without hardcoded @example.com)
	sessionData := models.SessionUser{
		ID: user.UserId,
	}

	// Generate session ID
	sessionId, err := utils.GenerateTokenSession()
	if err != nil {
		return utils.ResponseErrorModal(c, "Login Failed", "Failed to create session")
	}

	// Store session data
	middlewares.CreateSession(c, middlewares.SESSION_USER_ID, sessionData.ID)
	middlewares.CreateSession(c, middlewares.SESSION_ID, sessionId)

	return utils.ResponseRedirectHTMX(c, "/", fiber.StatusAccepted)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	time.Sleep(1 * time.Second) // Reduced for better UX
	middlewares.DeleteSession(c)

	// redirect to login page
	return utils.ResponseRedirectHTMX(c, "/login", fiber.StatusSeeOther)
}

// RegisterView shows the registration form
func (h *AuthHandler) RegisterView(c *fiber.Ctx) error {
	return adaptor.HTTPHandler(templ.Handler(components.RegisterPage()))(c)
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	time.Sleep(1 * time.Second) // Simulate processing time

	username := c.FormValue("username")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	// Validate input
	if username == "" || password == "" || confirmPassword == "" {
		return utils.ResponseErrorModal(c, "Registration Failed", "All fields are required")
	}

	if password != confirmPassword {
		return utils.ResponseErrorModal(c, "Registration Failed", "Passwords do not match")
	}

	if len(password) < 6 {
		return utils.ResponseErrorModal(c, "Registration Failed", "Password must be at least 6 characters long")
	}

	// Hash password
	hashedPassword, err := models.HashPassword(password)
	if err != nil {
		return utils.ResponseErrorModal(c, "Registration Failed", "Failed to process registration")
	}

	// Create user
	userData := models.UserCreate{
		Username: username,
		Password: hashedPassword,
	}

	// Check if username already exists and create user
	_, err = h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Check if user already exists
		_, err := h.usersRepo.FindByUsername(tx, username)
		if err == nil {
			return fiber.StatusConflict, fiber.NewError(fiber.StatusConflict, "Username already exists")
		}
		if err != sql.ErrNoRows {
			return fiber.StatusInternalServerError, err
		}

		// Create new user
		if err := h.usersRepo.Create(tx, userData); err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	})

	if err != nil {
		if err.Error() == "Username already exists" {
			return utils.ResponseErrorModal(c, "Registration Failed", "Username already exists")
		}
		return utils.ResponseErrorModal(c, "Registration Failed", "Failed to create user")
	}

	// Show success message and redirect to login
	return utils.ResponseSuccessModal(c, "Registration Successful", "Your account has been created. Please login.", false)
}
