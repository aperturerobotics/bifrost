package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggingMiddlewareOpts are opts passed to LoggingMiddleware.
type LoggingMiddlewareOpts struct {
	// UserAgent includes user agent in logs.
	UserAgent bool
}

// LoggingMiddleware logs incoming requests and response status codes using logrus.
func LoggingMiddleware(next http.Handler, le *logrus.Entry, opts LoggingMiddlewareOpts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to capture the status code
		wrappedWriter := &statusCapturingResponseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(wrappedWriter, r)

		// Log the request and response status code
		WithLoggerFields(le, r, wrappedWriter.statusCode).
			Debug("handled request")
	})
}

// WithLoggerFields builds the log fields for a request.
func WithLoggerFields(le *logrus.Entry, r *http.Request, status int) *logrus.Entry {
	fields := logrus.Fields{
		"method": r.Method,
		"uri":    r.RequestURI,
	}
	if userAgent := r.UserAgent(); userAgent != "" {
		fields["user-agent"] = userAgent
	}
	if status != 0 {
		fields["status"] = status
	}
	return le.WithFields(fields)
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}
