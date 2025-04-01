package httpServer

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/utils"
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidTokenFormat = errors.New("invalid token format")

// parseToken parses jwt token from header
func parseToken(authHeader string) (string, error) {
	// Parse token
	if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
		return "", ErrInvalidTokenFormat
	}
	tokenString := authHeader[len("Bearer "):]
	return tokenString, nil
}

// Middleware to protect routes
func protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Missing token")
			return
		}

		tokenString, err := parseToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token format")
			return

		}
		token, err := jwt.ParseWithClaims(tokenString, &utils.Claims{}, func(token *jwt.Token) (any, error) {
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
	// Define allowed origins
	allowedOrigins := map[string]bool{
		"http://localhost:8090": true,
		"http://localhost:8080": true,
		"http://localhost:8081": true,
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

func AllowAccess(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: in progres
		if !slices.Contains(roles, "") {
			c.AbortWithStatus(http.StatusForbidden)
		}
		c.Next()
	}
}
