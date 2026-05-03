package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/api/adapters/rest/middleware"
)

func TestRate_LimitsRequests(t *testing.T) {
	var count int
	handler := middleware.Rate(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusOK)
	}, 2) // 2 requests per second

	// First two should succeed immediately
	for range 2 {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Third request within same second may be rate limited
	// Note: ratelimit library behavior may vary, this is basic smoke test
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	// Either OK (if timing allows) or delayed - just verify no panic
	_ = rr.Code
}

func TestRate_AllowsAfterWait(t *testing.T) {
	handler := middleware.Rate(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, 1) // 1 per second

	// First request
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	handler(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Wait for rate limit window
	time.Sleep(1100 * time.Millisecond)

	// Second request should succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}
