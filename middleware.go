package slogchi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

var customAttributesCtxKey = customAttributesCtxKeyType{}

type customAttributesCtxKeyType struct{}

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

	WithUserAgent      bool
	WithRequestID      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool
	WithSpanID         bool
	WithTraceID        bool

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

		WithUserAgent:      false,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,
		WithSpanID:         false,
		WithTraceID:        false,

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

		WithUserAgent:      false,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,
		WithSpanID:         false,
		WithTraceID:        false,

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
			var (
				reqBody     []byte
				reqBodySize int
			)
			buf, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewBuffer(buf))
				if len(buf) > RequestBodyMaxSize {
					reqBody = buf[:RequestBodyMaxSize]
				} else {
					reqBody = buf
				}
				reqBodySize = len(buf)
			}

			// dump response body
			if config.WithResponseBody {
				w = newBodyWriter(w, ResponseBodyMaxSize)
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				status := ww.Status()
				method := r.Method
				route := chi.RouteContext(r.Context()).RoutePattern()
				end := time.Now()
				latency := end.Sub(start)
				userAgent := r.UserAgent()

				rqAttributes := []slog.Attr{
					slog.Time("time", end),
					slog.Duration("latency", latency),
					slog.String("method", method),
					slog.String("path", path),
					slog.String("route", route),
					slog.Int("status", status),
					slog.String("ip", r.RemoteAddr),
				}

				if config.WithUserAgent {
					rqAttributes = append(rqAttributes, slog.String("user-agent", userAgent))
				}

				if config.WithRequestID {
					rqAttributes = append(rqAttributes, slog.String("id", middleware.GetReqID(r.Context())))
				}

				// otel
				if config.WithTraceID {
					traceID := trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()
					rqAttributes = append(rqAttributes, slog.String("trace-id", traceID))
				}
				if config.WithSpanID {
					spanID := trace.SpanFromContext(r.Context()).SpanContext().SpanID().String()
					rqAttributes = append(rqAttributes, slog.String("span-id", spanID))
				}

				// request
				if config.WithRequestBody {
					rqAttributes = append(rqAttributes, slog.String("body", string(reqBody)))
				}

				rqAttributes = append(rqAttributes, slog.Int("bytes", reqBodySize))
				if config.WithRequestHeader {
					for k, v := range r.Header {
						if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
							continue
						}
						rqAttributes = append(rqAttributes, slog.Group("header", slog.Any(k, v)))
					}
				}

				var rsAttributes []slog.Attr
				// response
				if config.WithResponseBody {
					if w, ok := w.(*bodyWriter); ok {
						rsAttributes = append(rsAttributes, slog.String("body", w.body.String()))
					}
				}
				rsAttributes = append(rsAttributes, slog.Int("bytes", ww.BytesWritten()))
				if config.WithResponseHeader {
					for k, v := range w.Header() {
						if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
							continue
						}
						rsAttributes = append(rsAttributes, slog.Group("header", slog.Any(k, v)))
					}
				}

				attributes := []slog.Attr{
					{
						Key:   "request",
						Value: slog.GroupValue(rqAttributes...),
					},
					{
						Key:   "response",
						Value: slog.GroupValue(rsAttributes...),
					},
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
					level = config.ServerErrorLevel
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = config.ClientErrorLevel
				}

				logger.LogAttrs(r.Context(), level, strconv.Itoa(status)+": "+http.StatusText(status), attributes...)
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
