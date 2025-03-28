package httpServer

import (
	"ePrometna_Server/config"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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

	jwt, _, err := GenerateTokens("Test")
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
