package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/api/adapters/rest/middleware"
)

type mockVerifier struct {
	valid bool
	err   error
}

func (m *mockVerifier) Verify(token string) error {
	if m.valid {
		return nil
	}
	return m.err
}

func TestAuth_MissingHeader(t *testing.T) {
	handler := middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, &mockVerifier{valid: true})
	
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	handler(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing Authorization header")
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	handler := middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, &mockVerifier{valid: true})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()
	
	handler(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid Authorization header format")
}

func TestAuth_VerificationFailed(t *testing.T) {
	handler := middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, &mockVerifier{valid: false, err: assert.AnError})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token token123")
	rr := httptest.NewRecorder()
	
	handler(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid or expired token")
}

func TestAuth_Success(t *testing.T) {
	called := false
	handler := middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}, &mockVerifier{valid: true})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token validtoken")
	rr := httptest.NewRecorder()
	
	handler(rr, req)
	
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}
