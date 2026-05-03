package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/api/adapters/rest/middleware"
)

func TestConcurrency_WithinLimit(t *testing.T) {
	var count int
	var mu sync.Mutex
	handler := middleware.Concurrency(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}, 3)

	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			handler(rr, req)
		})
	}
	wg.Wait()

	assert.Equal(t, 3, count)
}

func TestConcurrency_ExceedLimit(t *testing.T) {
	handler := middleware.Concurrency(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}, 1)

	// First request
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	go handler(rr1, req1)

	// Give first request time to acquire semaphore
	time.Sleep(10 * time.Millisecond)

	// Second request should be rejected
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler(rr2, req2)

	assert.Equal(t, http.StatusServiceUnavailable, rr2.Code)
}
