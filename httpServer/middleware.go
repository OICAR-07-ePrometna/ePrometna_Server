package httpServer

import (
	"ePrometna_Server/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Middleware to protect routes
func protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetString("Authorization")
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
			return config.AppConfig.JwtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
			return
		}
	}
}

func corsHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: write hteaders if neaded
	}
}

/*
func headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("Access-Control-Allow-Origin", "http://localhost:3000")
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token, Access-Control-Allow-Origin")
		headers.Add("Access-Control-Allow-Methods", "GET,POST,OPTIONS,DELETE")
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
*/
