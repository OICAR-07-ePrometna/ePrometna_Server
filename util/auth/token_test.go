package auth_test

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Mock AppConfig for testing
var mockAppConfig = &config.AppConfiguration{
	JwtKey:     "test-jwt-key",
	RefreshKey: "test-refresh-key",
}

func TestParseToken(t *testing.T) {
	config.AppConfig = mockAppConfig // Use mock config for testing

	now := time.Now()
	accessTokenDuration := 5 * time.Minute
	validClaims := auth.Claims{
		Email: "test@example.com",
		Uuid:  "test-uuid",
		Role:  model.RoleOsoba,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
		},
	}
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
	validTokenString, _ := validToken.SignedString([]byte(mockAppConfig.JwtKey))
	validAuthHeader := "Bearer " + validTokenString

	invalidSignatureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
	invalidSignatureTokenString, _ := invalidSignatureToken.SignedString([]byte("wrong-key"))
	invalidSignatureAuthHeader := "Bearer " + invalidSignatureTokenString

	expiredClaims := auth.Claims{
		Email: "expired@example.com",
		Uuid:  "expired-uuid",
		Role:  model.UserRole("user"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(mockAppConfig.JwtKey))
	expiredAuthHeader := "Bearer " + expiredTokenString

	tests := []struct {
		name       string
		authHeader string
		wantToken  bool
		wantClaims *auth.Claims
		wantErr    error
	}{
		{
			name:       "Valid token",
			authHeader: validAuthHeader,
			wantToken:  true,
			wantClaims: &validClaims,
			wantErr:    nil,
		},
		{
			name:       "Invalid token format - missing Bearer",
			authHeader: validTokenString,
			wantToken:  false,
			wantClaims: nil,
			wantErr:    cerror.ErrInvalidTokenFormat,
		},
		{
			name:       "Invalid token format - too short",
			authHeader: "Bearer",
			wantToken:  false,
			wantClaims: nil,
			wantErr:    cerror.ErrInvalidTokenFormat,
		},
		{
			name:       "Invalid signature",
			authHeader: invalidSignatureAuthHeader,
			wantToken:  false,
			wantClaims: nil,
			wantErr:    jwt.ErrSignatureInvalid,
		},
		{
			name:       "Expired token",
			authHeader: expiredAuthHeader,
			wantToken:  false,
			wantClaims: nil,
			wantErr:    jwt.ErrTokenExpired,
		},
		{
			name:       "Malformed token",
			authHeader: "Bearer malformed.token.string",
			wantToken:  false,
			wantClaims: nil,
			wantErr:    jwt.ErrTokenMalformed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, gotClaims, gotErr := auth.ParseToken(tt.authHeader)

			if gotErr != nil {
				if tt.wantErr == nil || !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ParseToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr != nil {
				t.Errorf("ParseToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if (gotToken != nil) != tt.wantToken {
				t.Errorf("ParseToken() gotToken = %v, wantToken %v", gotToken != nil, tt.wantToken)
			}

			if tt.wantClaims != nil {
				if gotClaims == nil {
					t.Errorf("ParseToken() gotClaims = nil, want non-nil")
					return
				}
				if !reflect.DeepEqual(gotClaims, tt.wantClaims) {
					t.Errorf("ParseToken() gotClaims = %v, wantClaims %v", gotClaims, tt.wantClaims)
				}
			} else if gotClaims != nil {
				t.Errorf("ParseToken() gotClaims = %v, want nil", gotClaims)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	config.AppConfig = mockAppConfig // Use mock config for testing

	user := &model.User{
		Email: "test@example.com",
		Uuid:  uuid.New(),
		Role:  model.RoleAdmin,
	}

	accessTokenDuration := 5 * time.Minute
	refreshTokenDuration := 7 * 24 * time.Hour

	tests := []struct {
		name                     string
		user                     *model.User
		wantAccessTokenNonEmpty  bool
		wantRefreshTokenNonEmpty bool
		wantErr                  bool
	}{
		{
			name:                     "Valid user",
			user:                     user,
			wantAccessTokenNonEmpty:  true,
			wantRefreshTokenNonEmpty: true,
			wantErr:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccessToken, gotRefreshToken, err := auth.GenerateTokens(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantAccessTokenNonEmpty && gotAccessToken == "" {
				t.Errorf("GenerateTokens() gotAccessToken = %q, want non-empty", gotAccessToken)
			}
			if tt.wantRefreshTokenNonEmpty && gotRefreshToken == "" {
				t.Errorf("GenerateTokens() gotRefreshToken = %q, want non-empty", gotRefreshToken)
			}

			if gotAccessToken != "" {
				token, _, err := new(jwt.Parser).ParseUnverified(gotAccessToken, &auth.Claims{})
				if err != nil {
					t.Errorf("GenerateTokens() generated invalid access token: %v", err)
					return
				}
				if claims, ok := token.Claims.(*auth.Claims); ok {
					if claims.Email != tt.user.Email || claims.Uuid != tt.user.Uuid.String() || claims.Role != tt.user.Role {
						t.Errorf("GenerateTokens() access token claims are incorrect: got = %v, want = %v", claims, tt.user)
					}
					if !claims.ExpiresAt.After(time.Now().Add(accessTokenDuration-time.Minute)) || !claims.ExpiresAt.Before(time.Now().Add(accessTokenDuration+time.Minute)) {
						t.Errorf("GenerateTokens() access token expiry is not within expected range: %v", claims.ExpiresAt)
					}
				} else {
					t.Errorf("GenerateTokens() failed to parse access token claims")
				}
			}

			if gotRefreshToken != "" {
				token, _, err := new(jwt.Parser).ParseUnverified(gotRefreshToken, &auth.Claims{})
				if err != nil {
					t.Errorf("GenerateTokens() generated invalid refresh token: %v", err)
					return
				}
				if claims, ok := token.Claims.(*auth.Claims); ok {
					if claims.Email != tt.user.Email || claims.Uuid != tt.user.Uuid.String() || claims.Role != tt.user.Role {
						t.Errorf("GenerateTokens() refresh token claims are incorrect: got = %v, want = %v", claims, tt.user)
					}
					if !claims.ExpiresAt.After(time.Now().Add(refreshTokenDuration-time.Minute)) || !claims.ExpiresAt.Before(time.Now().Add(refreshTokenDuration+time.Minute)) {
						t.Errorf("GenerateTokens() refresh token expiry is not within expected range: %v", claims.ExpiresAt)
					}
				} else {
					t.Errorf("GenerateTokens() failed to parse refresh token claims")
				}
			}
		})
	}
}
