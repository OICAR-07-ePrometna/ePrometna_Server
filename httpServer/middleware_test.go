package httpServer

import (
	"ePrometna_Server/config"
	"fmt"
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
	router.Use(protect())
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return router
}

func TestGenerateTokens(t *testing.T) {
	// Setup
	// TODO: This may be dangerouts
	if err := os.Chdir("../"); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Faled to load config %+v", err)
	}

	router := startTESTServer()
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ping", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Faled to generate tokens %+v", err)
	}

	jwt, _, err := GenerateTokens("Test")
	if err != nil {
		t.Fatalf("Faled to generate tokens %+v", err)
	}
	// Add token bearer
	req.Header.Add("Authorization", "Bearer "+jwt)

	fmt.Printf("req: %+v\n", req)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		body, _ := w.Body.ReadBytes('\n')
		t.Fatalf("Returnd code %d, err = %+v", w.Code, string(body))
	}
}
