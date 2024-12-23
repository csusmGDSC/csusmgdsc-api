package routes

import (
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo) {
	e.POST("/events", handlers.InsertEventHandler)

	authGroup := e.Group("/auth")
	authGroup.POST("/register", handlers.RegisterUser)
	authGroup.POST("/login", handlers.LoginUser)
	authGroup.PATCH("/refresh", handlers.RefreshUser)
	authGroup.POST("/logout", handlers.LogoutUser)
	authGroup.POST("/logoutAll", handlers.LogoutAll, auth.AuthMiddleware)
	// authGroup.GET("/:provider/login", auth.OAuthLogin)
	// authGroup.GET("/:provider/callback", auth.OAuthCallback)

	// User routes can be used called from User or Admin
	apiUserGroup := e.Group("/user")
	apiUserGroup.Use(auth.AuthMiddleware)
	apiUserGroup.PUT("/update/:id", handlers.UpdateUser)
	apiUserGroup.DELETE("/delete/:id", handlers.DeleteUser)
}
