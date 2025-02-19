package routes

import (
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_handlers"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_middleware"
	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo, h *handlers.Handler) {
	e.GET("/users", h.GetUsersHandler) // supports pagination ?page=x&limit=y
	e.GET("/users/:id", h.GetUserByIDHandler)

	e.POST("/events", h.InsertEventHandler)
	e.GET("/events", h.GetEventsHandler) // supports pagination ?page=x&limit=y
	e.GET("/events/:id", h.GetEventByIDHandler)
	e.PUT("/events/:id", h.UpdateEventByID)
	e.DELETE("/events/:id", h.DeleteEventByID)

	e.POST("/events/:id/organizers/:userId", h.AddEventOrganizer)
	e.GET("/events/:id/organizers", h.GetEventOrganizers)
	e.GET("/users/:id/events", h.GetUserAssignedEvents)
	e.DELETE("/events/:id/organizers/:userId", h.DeleteOrganizerFromEvent)

	adminGroup := e.Group("/admin")
	adminGroup.Use(auth_middleware.AuthMiddleware)
	adminGroup.POST("/events", h.InsertEventHandler)

}

func InitOAuthRoutes(e *echo.Echo, h *auth_handlers.OAuthHandler) {
	authGroup := e.Group("/auth")
	authGroup.POST("/register", h.RegisterUser)
	authGroup.POST("/login", h.LoginUser)
	authGroup.PATCH("/refresh", h.RefreshUser)
	authGroup.POST("/logout", h.LogoutUser)
	authGroup.POST("/logoutAll", h.LogoutAll, auth_middleware.AuthMiddleware)
	authGroup.GET("/:provider/login", h.OAuthLogin)
	authGroup.GET("/:provider/callback", h.OAuthCallback)
	authGroup.PUT("/update/:id", h.UpdateUser, auth_middleware.AuthMiddleware)
	authGroup.DELETE("/delete/:id", h.DeleteUser, auth_middleware.AuthMiddleware)
	authGroup.GET("/me", h.GetUserByIDHandler, auth_middleware.AuthMiddleware)
}
