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

	Store = session.New(session.Config{
		Expiration:     7 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		// CookieName:     SESSION_APP_COOKIE_ID, // not using CookiNaame because we set it in KeyLookup and Cookiname is deprecated in latest fiber version
		KeyLookup: "cookie:" + SESSION_APP_COOKIE_ID,
	})

	log.Println("Session store initialized")

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

	// err, _ = m.dbService.Transaction(c.Context(), func(tx *sql.Tx) (error, int) {
	// 	sessData, err := m.sessionRepo.FindSession(tx, session_id.(string), userid.(int))
	// 	if err != nil {
	// 		return err, fiber.StatusInternalServerError
	// 	}

	// 	if sessData.Id == 0 && sessData.UserId == 0 && sessData.SessionId == "" {
	// 		return fiber.NewError(fiber.StatusUnauthorized, "Session not found"), fiber.StatusUnauthorized
	// 	}

	// 	userData, err := m.userRepo.FindByID(tx, userid.(int))
	// 	if err != nil {
	// 		return err, fiber.StatusInternalServerError
	// 	}

	// 	userSession = sso_models.UserSession{
	// 		Id:               userData.Id,
	// 		Username:         userData.Username,
	// 		CreditToken:      userData.CreditToken,
	// 		LastFirstLLMUsed: userData.LastFirstLLMUsed,
	// 	}

	// 	return nil, fiber.StatusOK
	// })

	if err != nil {
		return handleRedirectAuthMiddleware(c, LOGIN_PAGE_URL, true)
	}

	// ! dummy userSessionData
	userSession = models.SessionUser{
		ID:    userid.(int),
		Email: "admin@admin.com",
		Role:  "admin",
	}

	// set user data session to local session data for parsing it to main handlers
	c.Locals(SESSION_USER_NAME, userSession)

	return c.Next()
}
