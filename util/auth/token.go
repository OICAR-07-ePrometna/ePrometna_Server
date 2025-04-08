package auth

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	Email string         `json:"email"`
	Uuid  string         `json:"uuid"`
	Role  model.UserRole `json:"role"`
}

// Token expiry durations
const (
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

// ParseToken parses jwt token from header
func ParseToken(authHeader string) (*jwt.Token, *Claims, error) {
	// Parse token
	if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
		return nil, nil, cerror.ErrInvalidTokenFormat
	}
	tokenString := authHeader[len("Bearer "):]
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.AppConfig.JwtKey), nil
	})
	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil
}

// Generate JWT access and refresh tokens
func GenerateTokens(user *model.User) (string, string, error) {
	// Create access token

	accessTokenClaims := &Claims{
		Email: user.Email,
		Uuid:  user.Uuid.String(),
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.AppConfig.JwtKey))
	if err != nil {
		zap.S().Debugf("Failed to generate access token err = %+v", err)
		return "", "", err
	}

	// Create refresh token
	refreshTokenClaims := &Claims{
		Email: user.Email,
		Uuid:  user.Uuid.String(),
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenDuration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.AppConfig.RefreshKey))
	if err != nil {
		zap.S().Debugf("Failed to generate refresh token err = %+v", err)
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
