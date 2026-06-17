package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/middleware"
)

const testSecret = "test-secret-minimum-32-characters-ok"

func makeToken(secret string, subject string, expiresAt time.Time) string {
	claims := &jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return tok
}

func TestJWTChecker_ValidToken(t *testing.T) {
	userID := uuid.New()
	token := makeToken(testSecret, userID.String(), time.Now().Add(time.Hour))

	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-User-ID")
		if got != userID.String() {
			t.Errorf("X-User-ID = %q, want %q", got, userID.String())
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestJWTChecker_MissingHeader(t *testing.T) {
	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestJWTChecker_MalformedHeader(t *testing.T) {
	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer token")
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestJWTChecker_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	token := makeToken(testSecret, userID.String(), time.Now().Add(-time.Hour))

	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestJWTChecker_WrongSecret(t *testing.T) {
	userID := uuid.New()
	token := makeToken("other-secret-minimum-32-characters-ok", userID.String(), time.Now().Add(time.Hour))

	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestJWTChecker_InvalidSubject(t *testing.T) {
	token := makeToken(testSecret, "not-a-uuid", time.Now().Add(time.Hour))

	checker := middleware.NewJWTChecker(testSecret)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	checker.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
