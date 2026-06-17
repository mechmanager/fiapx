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

func makeToken(secret, subject string, expiresAt time.Time) string {
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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	checker.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-User-ID"); got != userID.String() {
			t.Errorf("X-User-ID = %q, want %q", got, userID.String())
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestJWTChecker_Unauthorized(t *testing.T) {
	checker := middleware.NewJWTChecker(testSecret)
	uid := uuid.New()

	cases := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"malformed header", "NotBearer token"},
		{"expired token", "Bearer " + makeToken(testSecret, uid.String(), time.Now().Add(-time.Hour))},
		{"wrong secret", "Bearer " + makeToken("other-secret-minimum-32-characters-ok", uid.String(), time.Now().Add(time.Hour))},
		{"invalid subject", "Bearer " + makeToken(testSecret, "not-a-uuid", time.Now().Add(time.Hour))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rr := httptest.NewRecorder()

			checker.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("handler should not be called")
			})).ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rr.Code)
			}
		})
	}
}
