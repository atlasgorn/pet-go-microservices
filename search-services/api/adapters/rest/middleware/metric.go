package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rw.code)
		url := r.URL.Path

		metricName := `http_request_duration_seconds{status="` + statusStr + `",url="` + url + `"}`
		histogram := metrics.GetOrCreateHistogram(metricName)
		histogram.Update(duration)
	})
}
