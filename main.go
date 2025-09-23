package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/momokii/go-rab-maker/backend/handlers"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/utils"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	// handlers
	authHandler := handlers.NewAuthHandler()

	// session for auth
	session := middlewares.NewSessionMiddleware()

	app := fiber.New(fiber.Config{
		ServerHeader:    "RAB Maker",
		AppName:         "RAB Maker v1.1.0",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				// Allow redirects to pass through - don't treat them as errors
				if code >= 300 && code < 400 {
					return err
				}

			}
			// handle actual error
			return utils.ResponseErrorModal(c, "Error "+string(rune(code)), err.Error())
		},
		DisableStartupMessage: true,
	})

	middlewares.SetupAppMiddlewares(app)

	app.Get("/login", session.IsNotAuth, authHandler.LoginView)
	app.Post("/login", session.IsNotAuth, authHandler.Login)
	app.Post("/logout", session.IsAuth, authHandler.Logout)

	app.Get("/", session.IsAuth, handlers.CountriesTableView)
	app.Get("/countries/new", session.IsAuth, handlers.CreateCountriesModal)
	app.Get("/countries/:id/edit", session.IsAuth, handlers.EditCountriesModal)
	app.Post("/countries/:id/edit", session.IsAuth, handlers.EditCountries)
	app.Get("/countries/:id/delete", session.IsAuth, handlers.DeleteCountriesModal)
	app.Delete("/countries/:id/delete", session.IsAuth, handlers.DeleteCountries)
	app.Post("/countries/new", session.IsAuth, handlers.AddNewCountries)
	app.Get("/countries", session.IsAuth, handlers.CountriesView)

	app.Get("/countries/:id", handlers.GetCountryDetails)
	app.Get("/search", handlers.SearchCountries)

	app.Get("/demo/loading-test", handlers.DemoLoadingTest)
	app.Get("/demo/success-test", handlers.SucceessModalTest)
	app.Get("/demo/error-test", handlers.ErrorModalTest)

	startServerWithGracefulShutdown(app)
}

func startServerWithGracefulShutdown(app *fiber.App) {
	// channel for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // default port
	}
	go func() {
		log.Println("Starting server on port " + port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server startup error: %v", err)
		}
	}()

	log.Println("Server started successfully.")

	// Wait for shutdown signal
	<-quit
	log.Println("Shutdown signal received...")

	// Wait 5 seconds before starting graceful shutdown
	log.Println("Waiting 5 seconds before graceful shutdown...")
	time.Sleep(5 * time.Second)

	// Graceful shutdown
	log.Println("Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown due to error: %v", err)
	} else {
		log.Println("Server shutdown complete")
	}

	log.Println("Application terminated")
}
