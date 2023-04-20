package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggingMiddleware logs incoming requests and response status codes using logrus.
func LoggingMiddleware(next http.Handler, le *logrus.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to capture the status code
		wrappedWriter := &statusCapturingResponseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(wrappedWriter, r)

		// Log the request and response status code
		le.WithFields(logrus.Fields{
			"method": r.Method,
			"uri":    r.RequestURI,
			"status": wrappedWriter.statusCode,
		}).Info("handled request")
	})
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}
