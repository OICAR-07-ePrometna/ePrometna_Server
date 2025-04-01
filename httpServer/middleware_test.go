package httpServer

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
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

var once = sync.Once{}

func LoadConfigForTests() {
	once.Do(func() {
		// NOTE: Change directory
		if err := os.Chdir("../"); err != nil {
			panic(fmt.Sprintf("Failed to change directory: %v", err))
		}

		err := config.LoadConfig()
		if err != nil {
			panic(fmt.Sprintf("Failed to load config %+v", err))
		}
	})
}

func TestGenerateTokens(t *testing.T) {
	// Setup
	LoadConfigForTests()

	router := startTESTServer()
	router.Use(protect())

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
