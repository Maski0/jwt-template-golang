package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization Header Provided"})
		}
	}
}
