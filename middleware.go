package slogchi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	customAttributesCtxKey = "slog-chi.custom-attributes"
)

var (
	RequestBodyMaxSize  = 64 * 1024 // 64KB
	ResponseBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
		"authorization": {},
		"cookie":        {},
		"set-cookie":    {},
		"x-auth-token":  {},
		"x-csrf-token":  {},
		"x-xsrf-token":  {},
	}
	HiddenResponseHeaders = map[string]struct{}{
		"set-cookie": {},
	}
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool

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

		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

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

		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

		Filters: filters,
	})
}

// NewWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path

			// dump request body
			var reqBody []byte
			if config.WithRequestBody {
				buf, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewBuffer(buf))
					if len(buf) > RequestBodyMaxSize {
						reqBody = buf[:RequestBodyMaxSize]
					} else {
						reqBody = buf
					}
				}
			}

			// dump response body
			if config.WithResponseBody {
				w = newBodyWriter(w, ResponseBodyMaxSize)
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				// headers := ww.Header()
				// length := ww.BytesWritten()

				// status := r.Response.StatusCode
				status := ww.Status()
				method := r.Method
				route := chi.RouteContext(r.Context()).RoutePattern()
				end := time.Now()
				latency := end.Sub(start)
				// ip := "x.x.x.x"
				userAgent := r.UserAgent()

				attributes := []slog.Attr{
					slog.Time("time", end),
					slog.Duration("latency", latency),
					slog.String("method", method),
					slog.String("path", path),
					slog.String("route", route),
					slog.Int("status", status),
					// slog.String("ip", ip),
					slog.String("user-agent", userAgent),
				}

				// request
				if config.WithRequestBody {
					attributes = append(attributes, slog.Group("request", slog.String("body", string(reqBody))))
				}
				if config.WithRequestHeader {
					for k, v := range r.Header {
						if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
							continue
						}
						attributes = append(attributes, slog.Group("request", slog.Group("header", slog.Any(k, v))))
					}
				}

				// response
				if config.WithResponseBody {
					if w, ok := w.(*bodyWriter); ok {
						attributes = append(attributes, slog.Group("response", slog.String("body", w.body.String())))
					}
				}
				if config.WithResponseHeader {
					for k, v := range w.Header() {
						if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
							continue
						}
						attributes = append(attributes, slog.Group("response", slog.Group("header", slog.Any(k, v))))
					}
				}

				// custom context values
				if v := r.Context().Value(customAttributesCtxKey); v != nil {
					switch attrs := v.(type) {
					case []slog.Attr:
						attributes = append(attributes, attrs...)
					}
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

func AddCustomAttributes(r *http.Request, attr slog.Attr) {
	v := r.Context().Value(customAttributesCtxKey)
	if v == nil {
		*r = *r.WithContext(context.WithValue(r.Context(), customAttributesCtxKey, []slog.Attr{attr}))
		return
	}

	switch attrs := v.(type) {
	case []slog.Attr:
		*r = *r.WithContext(context.WithValue(r.Context(), customAttributesCtxKey, append(attrs, attr)))
	}
}
