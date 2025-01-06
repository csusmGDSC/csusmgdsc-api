package routes

import (
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo, h *handlers.Handler) {
	e.POST("/events", h.InsertEventHandler)

	authGroup := e.Group("/auth")
	authGroup.POST("/register", h.RegisterUser)
	authGroup.POST("/login", h.LoginUser)
	authGroup.PATCH("/refresh", h.RefreshUser)
	authGroup.POST("/logout", h.LogoutUser)
	authGroup.POST("/logoutAll", h.LogoutAll, auth.AuthMiddleware)
	// authGroup.GET("/:provider/login", auth.OAuthLogin)
	// authGroup.GET("/:provider/callback", auth.OAuthCallback)

	// User routes can be used called from User or Admin
	apiUserGroup := e.Group("/user")
	apiUserGroup.Use(auth.AuthMiddleware)
	apiUserGroup.PUT("/update/:id", h.UpdateUser)
	apiUserGroup.DELETE("/delete/:id", h.DeleteUser)
}
