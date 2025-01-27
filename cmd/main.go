package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_handlers"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/csusmGDSC/csusmgdsc-api/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	dbConn := db.GetInstance()
	defer dbConn.Close()

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")

	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:8081" // Default frontend origin
	}

	e := echo.New()

	// Initialize OAuth
	auth.InitOAuth()
	authHandler := auth_handlers.NewOAuthHandler(dbConn)
	routes.InitOAuthRoutes(e, authHandler)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Added CORS for frontend, need to allow frontend origin otherwise requests get rejected
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontendOrigin}, // Explicitly allow frontend origin
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Authorization"},
		AllowCredentials: true, // Important: Allow credentials to send over the cookie
	}))

	// Intialize API routes
	h := handlers.NewHandler(dbConn)
	routes.InitRoutes(e, h)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down server: %v", err)
		}
	}()

	log.Println("Starting server on port 8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
