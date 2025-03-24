package httpServer

import (
	"ePrometna_Server/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Middleware to protect routes
func protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Missing token")
			return
		}

		// Parse token
		if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token format")
			return
		}
		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
			return []byte(config.AppConfig.JwtKey), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
			return
		}
		c.Next()
	}
}

func corsHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: write hteaders if neaded
	}
}
