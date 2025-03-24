package httpServer

import (
	"ePrometna_Server/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Uuid     string `json:"uuid"`
	Role     string `json:"role"`
}

// Token expiry durations
const (
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

// Generate JWT access and refresh tokens
// TODO: Send Whole user struct
func GenerateTokens(username string) (string, string, error) {
	// Create access token

	accessTokenClaims := &Claims{
		// TODO: register Uuid and Role
		Username: username,
		Uuid:     "",
		Role:     "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.AppConfig.JwtKey))
	if err != nil {
		return "", "", err
	}

	// Create refresh token
	refreshTokenClaims := &Claims{
		// TODO: register Uuid and Role
		Username: username,
		Uuid:     "",
		Role:     "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenDuration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.AppConfig.RefreshKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
