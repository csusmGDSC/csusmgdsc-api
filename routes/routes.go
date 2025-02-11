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
	e.GET("/events", h.GetEventsHandler) // supports pagination ?page=x&limit=y
	e.GET("/events/:id", h.GetEventByIDHandler)

	adminGroup := e.Group("/admin")
	adminGroup.Use(auth_middleware.AuthMiddleware)
	adminGroup.POST("/events", h.InsertEventHandler)

	// Comments
	e.POST("/comments", h.InsertCommentHandler, auth_middleware.AuthMiddleware)       // Create comment (Requires auth)
	e.GET("/comments", h.GetCommentsHandler)                                          // Get comments by ?event_id=x&user_id=y
	e.GET("/comments/:id/replies", h.GetCommentRepliesHandler)                        // Get replies for a comment by id
	e.GET("/comments/:id", h.GetCommentByIdHandler)                                   // Get comment by id
	e.PUT("/comments/:id", h.UpdateCommentHandler, auth_middleware.AuthMiddleware)    // Update comment by id (Requires auth)
	e.DELETE("/comments/:id", h.DeleteCommentHandler, auth_middleware.AuthMiddleware) // Delete comment by id (Requires auth)
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
