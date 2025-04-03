package utils

import (
	"ePrometna_Server/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func HandleRefresh(c *gin.Context) {
	// Extract the refresh token from the request body
	refreshTokenString := c.PostForm("refresh_token")

	// Parse and verify the refresh token
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrAbortHandler
		}
		return []byte(config.AppConfig.RefreshKey), nil
	})

	// Handle invalid or expired refresh tokens
	if err != nil || !refreshToken.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Extract claims and validate the refresh token
	claims, ok := refreshToken.Claims.(*Claims)
	if !ok || claims.ExpiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Generate a new access token
	newAccessTokenClaims := &Claims{
		Email: claims.Email,
		Uuid:  claims.Uuid,
		Role:  claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)), // Access token expires in 5 minutes
		},
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessTokenClaims)

	// Sign the new access token
	newAccessTokenString, err := newAccessToken.SignedString([]byte(config.AppConfig.JwtKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	// Return the new access token
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessTokenString,
	})
}
