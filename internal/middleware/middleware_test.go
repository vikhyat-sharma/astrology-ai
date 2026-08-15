package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// stubValidator is a minimal TokenValidator for testing.
type stubValidator struct {
	id  uuid.UUID
	err error
}

func (s *stubValidator) ValidateToken(_ string) (uuid.UUID, error) {
	return s.id, s.err
}

func TestCORSAllowedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatal("expected CORS header for allowed origin")
	}
}

func TestCORSBlockedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header for blocked origin")
	}
}

func TestCORSPreflight(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.OPTIONS("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestAuthRequiredMissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(AuthRequired(&stubValidator{}))
	router.GET("/protected", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequiredInvalidFormat(t *testing.T) {
	router := gin.New()
	router.Use(AuthRequired(&stubValidator{}))
	router.GET("/protected", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequiredInvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthRequired(&stubValidator{err: errors.New("invalid")}))
	router.GET("/protected", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequiredValidToken(t *testing.T) {
	id := uuid.New()
	router := gin.New()
	router.Use(AuthRequired(&stubValidator{id: id}))
	router.GET("/protected", func(c *gin.Context) {
		got, _ := c.Get("userID")
		if got.(uuid.UUID) != id {
			t.Errorf("expected userID %v, got %v", id, got)
		}
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequestID(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(RateLimit(1, 1)) // 1 rps, burst 1
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	// First request should pass.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 on first request, got %d", w.Code)
	}

	// Second immediate request should be rate-limited.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on second request, got %d", w.Code)
	}
}

func TestLoggerMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) { panic("test panic") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}
}
