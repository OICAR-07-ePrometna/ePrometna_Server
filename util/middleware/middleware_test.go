package middleware_test

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// --- Test Suite Definition ---
type MiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
	sugar  *zap.SugaredLogger
}

// SetupSuite runs once before all tests in the suite
func (suite *MiddlewareTestSuite) SetupSuite() {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	zapLogger, _ := loggerCfg.Build()
	suite.sugar = zapLogger.Sugar()

	config.AppConfig = &config.AppConfiguration{
		AccessKey:     "test-middleware-access-key",
		RefreshKey:    "test-middleware-refresh-key",
		IsDevelopment: true,
		Port:          8090, // Not directly used by middleware tests
		DbConnection:  "",   // Not used by middleware tests
	}

	// Setup Gin router for testing
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	suite.router.GET("/protected/general", middleware.Protect(), func(c *gin.Context) {
		c.String(http.StatusOK, "general_access_granted")
	})

	suite.router.GET("/protected/admin", middleware.Protect(model.RoleSuperAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "admin_access_granted")
	})

	suite.router.GET("/cors-test", middleware.CorsHeader(), func(c *gin.Context) {
		c.String(http.StatusOK, "cors_ok")
	})
	suite.router.OPTIONS("/cors-test", middleware.CorsHeader(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// TearDownSuite runs once after all tests in the suite
func (suite *MiddlewareTestSuite) TearDownSuite() {
	if suite.sugar != nil {
		suite.sugar.Sync()
	}
}

// Helper to make HTTP requests
func (suite *MiddlewareTestSuite) performRequest(method, path, token string, body ...string) *httptest.ResponseRecorder {
	var reqBody *strings.Reader
	if len(body) > 0 {
		reqBody = strings.NewReader(body[0])
	} else {
		reqBody = strings.NewReader("")
	}

	req, _ := http.NewRequest(method, path, reqBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// Helper to generate a token
func (suite *MiddlewareTestSuite) generateToken(userID uuid.UUID, userEmail string, userRole model.UserRole, expiresAt time.Time) string {
	claims := &auth.Claims{
		Email: userEmail,
		Uuid:  userID.String(),
		Role:  userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), // Allow for slight clock skew
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.AccessKey))
	if err != nil {
		suite.T().Fatalf("Failed to sign token for test: %v", err)
	}
	return tokenString
}

// --- Test Cases for Protect Middleware ---

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_GeneralAccess() {
	testUserUUID := uuid.New()
	token := suite.generateToken(testUserUUID, "test@example.com", model.RoleOsoba, time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/general", token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "general_access_granted", w.Body.String())
}

func (suite *MiddlewareTestSuite) TestProtect_NoToken() {
	w := suite.performRequest(http.MethodGet, "/protected/general", "")

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Missing token")
}

func (suite *MiddlewareTestSuite) TestProtect_InvalidTokenFormat_NoBearer() {
	w := httptest.NewRecorder() // Use httptest directly for more control over header
	req, _ := http.NewRequest(http.MethodGet, "/protected/general", nil)
	req.Header.Set("Authorization", "InvalidTokenWithoutBearerPrefix")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token format")
}

func (suite *MiddlewareTestSuite) TestProtect_InvalidTokenFormat_TooShort() {
	w := suite.performRequest(http.MethodGet, "/protected/general", "short")

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token format")
}

func (suite *MiddlewareTestSuite) TestProtect_MalformedToken() {
	// This token is not a valid JWT structure
	malformedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" // Missing signature part
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected/general", nil)
	req.Header.Set("Authorization", malformedToken) // Set directly to bypass "Bearer " prefixing in helper for this specific case
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// The error message might vary based on JWT library, "Invalid token" or "token contains an invalid number of segments"
	// For this test, we check for "Invalid token" as per the middleware's generic error for parsing issues.
	// A more specific check might involve parsing the JSON response if your middleware returns one.
	// The current middleware returns "Invalid token" for jwt.ParseWithClaims errors.
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_ExpiredToken() {
	testUserUUID := uuid.New()
	expiredToken := suite.generateToken(testUserUUID, "expired@example.com", model.RoleOsoba, time.Now().Add(-5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/general", expiredToken)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// The middleware uses `token.Valid` which would be false for an expired token.
	// The specific error from `jwt.ParseWithClaims` would be `jwt.ErrTokenExpired`.
	// The middleware then returns "Invalid token".
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_WrongSigningKey() {
	claims := &auth.Claims{
		Email: "wrongkey@example.com",
		Uuid:  uuid.New().String(),
		Role:  model.RoleOsoba,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with a different key
	wrongKeyToken, _ := token.SignedString([]byte("a-completely-different-secret-key"))

	w := suite.performRequest(http.MethodGet, "/protected/general", wrongKeyToken)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// Error from `jwt.ParseWithClaims` would be `jwt.ErrSignatureInvalid`.
	// Middleware returns "Invalid token".
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_CorrectRole() {
	testUserUUID := uuid.New()
	adminToken := suite.generateToken(testUserUUID, "admin@example.com", model.RoleSuperAdmin, time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/admin", adminToken)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "admin_access_granted", w.Body.String())
}

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_InsufficientRole() {
	testUserUUID := uuid.New()
	userToken := suite.generateToken(testUserUUID, "user@example.com", model.RoleOsoba, time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/admin", userToken)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	// No body expected for 403 from this middleware implementation
}

// --- Test Cases for CorsHeader Middleware ---
func (suite *MiddlewareTestSuite) TestCorsHeader_AllowsConfiguredOrigin() {
	allowedOrigin := "http://localhost:3000" // Must match one in your CorsHeader middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/cors-test", nil)
	req.Header.Set("Origin", allowedOrigin)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), allowedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.NotEmpty(suite.T(), w.Header().Get("Access-Control-Allow-Headers"))
	assert.NotEmpty(suite.T(), w.Header().Get("Access-Control-Allow-Methods"))
}

func (suite *MiddlewareTestSuite) TestCorsHeader_DoesNotAllowUnconfiguredOrigin() {
	unallowedOrigin := "http://untrusted-site.com"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/cors-test", nil)
	req.Header.Set("Origin", unallowedOrigin)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// If origin is not in allowedOrigins, Access-Control-Allow-Origin should not be set to the unallowedOrigin
	// It might be empty or not present, depending on default browser/server behavior if not explicitly set.
	// The current middleware implementation only sets it if origin is in allowedOrigins.
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Origin"), "Access-Control-Allow-Origin should be empty for unallowed origins")
}

func (suite *MiddlewareTestSuite) TestCorsHeader_OptionsPreflight() {
	allowedOrigin := "http://localhost:8081" // Must match one in your CorsHeader middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/cors-test", nil)
	req.Header.Set("Origin", allowedOrigin)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "authorization")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), allowedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Authorization") // Case-insensitive check
	assert.Equal(suite.T(), "86400", w.Header().Get("Access-Control-Max-Age"))
}

// --- Run Test Suite ---
func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
