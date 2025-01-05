package routes

import (
	"github.com/Maski0/jwt-template-golang/controllers"
	"github.com/Maski0/jwt-template-golang/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	userGroup := incomingRoutes.Group("/users")
	userGroup.Use(middleware.Authenticate())
	{
		userGroup.GET("/", controllers.GetUsers())
		userGroup.GET("/:user_id", controllers.GetUser())
	}
}
