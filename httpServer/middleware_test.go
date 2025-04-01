package httpServer

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/utils"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// startTESTServer is a test function
func startTESTServer() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return router
}

func TestGenerateTokens(t *testing.T) {
	// Setup
	// TODO: This may be dangerous
	if err := os.Chdir("../"); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config %+v", err)
	}

	router := startTESTServer()
	router.Use(protect())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ping", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Failed to generate tokens %+v", err)
	}

	jwt, _, err := utils.GenerateTokens(model.User{Email: "Test@test.t", Uuid: uuid.New()})
	if err != nil {
		t.Fatalf("Failed to generate tokens %+v", err)
	}
	// Add token bearer
	req.Header.Add("Authorization", "Bearer "+jwt)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Returned code %d, err = %+v", w.Code, w.Body.String())
	} else {
		expected := "pong"
		if w.Body.String() != expected {
			t.Fatalf("Expected body: %q, got: %q", expected, w.Body.String())
		}
	}
}

// Helper function to create a test JWT
func createTestJWT(t *testing.T, claims utils.Claims) string {
	// Create a simple header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Convert header to JSON
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("Failed to marshal header: %v", err)
	}

	// Convert claims to JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("Failed to marshal claims: %v", err)
	}

	// Base64 encode header
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Base64 encode claims
	claimsBase64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create a fake signature (doesn't matter for these tests)
	fakeSig := "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	// Combine to form token
	return headerBase64 + "." + claimsBase64 + "." + fakeSig
}
