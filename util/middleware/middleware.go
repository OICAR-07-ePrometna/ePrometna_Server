package middleware

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

var OptionsHandler gin.HandlerFunc = func(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Protect protects routes allowing access only to given roles (model.UserRole)
// if roles are empty they it only checks for the validity of tokens
func Protect(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Missing token")
			return
		}

		token, claims, err := auth.ParseToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token format")
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
			return
		}

		if len(roles) != 0 && !slices.Contains(roles, claims.Role) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func CorsHeader() gin.HandlerFunc {
	// Define allowed origins
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:8081": true,
		"http://localhost:8082": true,
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
