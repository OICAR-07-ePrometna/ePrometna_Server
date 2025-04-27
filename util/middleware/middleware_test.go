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

// Mock AppConfig for testing
var mockAppConfig = &config.AppConfiguration{
	AccessKey:     "test-jwt-key",
	RefreshKey:    "test-refresh-key",
	IsDevelopment: true,
	Port:          8090,
	DbConnection:  "",
}

func Setup() {
	config.AppConfig = mockAppConfig
}

func TestGenerateTokens(t *testing.T) {
	Setup()

	router := startTESTServer()
	router.Use(middleware.Protect())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ping", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Failed to generate tokens %+v", err)
	}

	jwt, _, err := auth.GenerateTokens(&model.User{Email: "Test@test.t", Uuid: uuid.New()})
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
