package middlewares

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/utils"
)

var (
	Store          *session.Store
	SECRET_KEY_JWT = os.Getenv("SECRET_KEY_JWT")
)

const (
	SESSION_APP_COOKIE_ID = "session_id_rabmaker"
	SESSION_USER_ID       = "id"
	SESSION_ID            = "session_id"
	SESSION_USER_NAME     = "user"

	LOGIN_PAGE_URL = "/login"
	DASHBOARD_URL  = "/"
)

type SessionMiddleware struct{}

func NewSessionMiddleware() *SessionMiddleware {
	// Check if running in production environment
	// Default to development mode if ENV is not set
	isProduction := os.Getenv("ENV") == "production"

	Store = session.New(session.Config{
		Expiration:     7 * time.Hour,
		CookieSecure:   isProduction, // Only use Secure cookies on HTTPS (production)
		CookieHTTPOnly: true,
		// CookieName:     SESSION_APP_COOKIE_ID, // not using CookiNaame because we set it in KeyLookup and Cookiname is deprecated in latest fiber version
		KeyLookup: "cookie:" + SESSION_APP_COOKIE_ID,
	})

	log.Println("Session store initialized (production mode: ", isProduction, ")")

	return &SessionMiddleware{}
}

func CreateSession(c *fiber.Ctx, key string, value interface{}) error {
	sess, err := Store.Get(c)
	if err != nil {
		return err
	}
	defer sess.Save()

	sess.Set(key, value)

	return nil
}

func DeleteSession(c *fiber.Ctx) error {
	sess, err := Store.Get(c)
	if err != nil {
		return err
	}
	defer sess.Save()

	sess.Destroy()

	return nil
}

func CheckSession(c *fiber.Ctx, key string) (interface{}, error) {
	sess, err := Store.Get(c)
	if err != nil {
		return nil, err
	}

	return sess.Get(key), nil
}

func handleRedirectAuthMiddleware(c *fiber.Ctx, url string, deleteSession bool) error {
	if deleteSession {
		DeleteSession(c)
	}

	// check if the request from HTMX or not
	if c.Get("HX-Request") == "true" {
		// if HTMX, send  header HX-Redirect.
		// main point is to not send any other body/header swap.
		return utils.ResponseRedirectHTMX(c, url, fiber.StatusUnauthorized)
	}

	// if request come for usual browser, so do normal redirect
	return c.Redirect(url, fiber.StatusFound) // StatusFound (302) for standard  redirect
}

func (m *SessionMiddleware) IsNotAuth(c *fiber.Ctx) error {

	userid, err := CheckSession(c, SESSION_USER_ID)
	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)

	}

	session_id, err := CheckSession(c, SESSION_ID)
	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	// user is still valid, so we redirect to home page
	if userid != nil && session_id != nil {
		return handleRedirectAuthMiddleware(c, DASHBOARD_URL, false)
	}

	return c.Next()
}

func (m *SessionMiddleware) GetUserSessionData(c *fiber.Ctx) (models.SessionUser, error) {
	var sessionData models.SessionUser

	userData := c.Locals(SESSION_USER_NAME).(models.SessionUser)

	if userData.ID == 0 {
		return sessionData, fiber.NewError(fiber.StatusUnauthorized, "Session not found")
	}

	sessionData = userData

	return sessionData, nil
}

// The function `IsAuth` in Go validates a JWT token from the Authorization header in a Fiber context.
func (m *SessionMiddleware) IsAuth(c *fiber.Ctx) error {

	userid, err := CheckSession(c, SESSION_USER_ID)
	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	session_id, err := CheckSession(c, SESSION_ID)
	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	if userid == nil || session_id == nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	var userSession models.SessionUser

	// TODO: Fetch real user data from database instead of using dummy data
	// The commented code below shows the original intent, but it requires:
	// 1. Adding dbService and userRepo to SessionMiddleware struct
	// 2. Properly handling the users repository
	// For now, we use the userid from session which is sufficient for authorization
	//
	// err, _ = m.dbService.Transaction(c.Context(), func(tx *sql.Tx) (error, int) {
	//     usersRepo := users.NewUsersRepo()
	//     userData, err := usersRepo.FindByID(tx, userid.(int))
	//     if err != nil {
	//         return err, fiber.StatusInternalServerError
	//     }
	//     userSession = models.SessionUser{
	//         ID:    userData.Id,
	//         Email: userData.Email,
	//         Role:  userData.Role,
	//     }
	//     return nil, fiber.StatusOK
	// })

	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	// Temporary: Use session userid only
	// In production, fetch full user data from database
	userSession = models.SessionUser{
		ID: userid.(int),
		// Email and Role should be fetched from database
		// Leaving these empty for now as they are not currently used
	}

	// set user data session to local session data for parsing it to main handlers
	c.Locals(SESSION_USER_NAME, userSession)

	return c.Next()
}
