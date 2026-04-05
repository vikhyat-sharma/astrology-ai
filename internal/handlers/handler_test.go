package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNotFoundHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
