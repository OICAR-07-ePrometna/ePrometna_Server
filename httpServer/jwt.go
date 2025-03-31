package httpServer

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Email string         `json:"username"`
	Uuid  string         `json:"uuid"`
	Role  model.UserRole `json:"role"`
}

// Token expiry durations
const (
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

// Generate JWT access and refresh tokens
func GenerateTokens(user model.User) (string, string, error) {
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
		return "", "", err
	}

	// Create refresh token
	refreshTokenClaims := &Claims{
		// TODO: register Uuid and Role
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
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// DecodeClaims decodes the claims portion of a JWT token
func DecodeClaims(tokenString string) (*Claims, error) {
	// Split the token into its three parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("token format invalid: not a JWT")
	}

	// Decode the claims part (the second part)
	claimsPart := parts[1]

	// Add padding if needed
	if l := len(claimsPart) % 4; l > 0 {
		claimsPart += strings.Repeat("=", 4-l)
	}

	// Decode base64
	claimsBytes, err := base64.URLEncoding.DecodeString(claimsPart)
	if err != nil {
		// Try with RawURLEncoding (no padding)
		claimsBytes, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("error decoding claims: %v", err)
		}
	}

	// Parse the claims
	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("error parsing claims: %v", err)
	}

	// Also parse into a map to capture custom claims
	var customClaims map[string]any
	if err := json.Unmarshal(claimsBytes, &customClaims); err != nil {
		return nil, fmt.Errorf("error parsing custom claims: %v", err)
	}

	return &claims, nil
}
