package slogchi

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID bool

	Filters []Filter
}

// New returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func New(logger *slog.Logger) func(http.Handler) http.Handler {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,

		Filters: []Filter{},
	})
}

// NewWithFilters returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewWithFilters(logger *slog.Logger, filters ...Filter) func(http.Handler) http.Handler {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,

		Filters: filters,
	})
}

// NewWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				// headers := ww.Header()
				// length := ww.BytesWritten()

				// status := r.Response.StatusCode
				status := ww.Status()
				method := r.Method
				end := time.Now()
				latency := end.Sub(start)
				// ip := "x.x.x.x"
				userAgent := r.UserAgent()

				attributes := []slog.Attr{
					slog.Time("time", end),
					slog.Duration("latency", latency),
					slog.String("method", method),
					slog.String("path", path),
					slog.Int("status", status),
					// slog.String("ip", ip),
					slog.String("user-agent", userAgent),
				}

				for _, filter := range config.Filters {
					if !filter(ww, r) {
						return
					}
				}

				level := config.DefaultLevel
				if status >= http.StatusInternalServerError {
					logger.LogAttrs(context.Background(), config.ServerErrorLevel, http.StatusText(status), attributes...)
					level = config.ServerErrorLevel
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = config.ClientErrorLevel
					logger.LogAttrs(context.Background(), config.ClientErrorLevel, http.StatusText(status), attributes...)
				}

				logger.LogAttrs(context.Background(), level, http.StatusText(status), attributes...)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
