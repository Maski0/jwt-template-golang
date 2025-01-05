// Base file
package main

import (
	"log"
	"os"

	"github.com/Maski0/jwt-template-golang/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading .env file")
	}
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}
	router := gin.Default()
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access Granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access Granted for api-2"})
	})

	router.Run(":" + port)
}
