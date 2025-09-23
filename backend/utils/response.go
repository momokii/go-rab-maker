package utils

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/frontend/components"
)

func ResponseErrorModal(c *fiber.Ctx, title, message string) error {
	c.Status(fiber.StatusUnprocessableEntity) // Set status code to 422
	c.Set("HX-Reswap", "none")                // Prevent swap if error occurs
	errorComponentModal := components.HTMXModalError(title, message)
	return adaptor.HTTPHandler(templ.Handler(errorComponentModal))(c)
}

func ResponseErrorModalNotAuthDirect(c *fiber.Ctx, title, message string) error {
	errorModal := components.HTMXModalErrorNotAuthDirect(title, message)
	return adaptor.HTTPHandler(templ.Handler(errorModal))(c)
}

func ResponseSuccessModal(c *fiber.Ctx, title, message string, preventSwap bool) error {
	successComponentModal := components.HTMXModalSuccess(title, message)

	if preventSwap {
		c.Set("HX-Reswap", "none") // Prevent swap if needed
	}

	return adaptor.HTTPHandler(templ.Handler(successComponentModal))(c)
}

func ResponseSuccessModalDirect(c *fiber.Ctx, title, message string) error {
	successComponentModal := components.HTMXModalSuccessDirect(title, message)
	return adaptor.HTTPHandler(templ.Handler(successComponentModal))(c)
}

func ResponseRedirectHTMX(c *fiber.Ctx, url string, status int) error {
	// using htmx for redirect required to add the URL at header
	c.Set("HX-Redirect", url)
	return c.SendStatus(status)
}
