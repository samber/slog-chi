package slogchi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type customAttributesCtxKeyType struct{}

var customAttributesCtxKey = customAttributesCtxKeyType{}

var (
	TraceIDKey   = "trace-id"
	SpanIDKey    = "span-id"
	RequestIDKey = "id"

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
			query := r.URL.RawQuery

			// dump request body
			br := newBodyReader(r.Body, RequestBodyMaxSize, config.WithRequestBody)
			r.Body = br

			// dump response body
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var bw *bodyWriter
			if config.WithResponseBody {
				bw = newBodyWriter(ResponseBodyMaxSize)
				ww.Tee(bw)
			}

			defer func() {
				params := map[string]string{}
				for i, k := range chi.RouteContext(r.Context()).URLParams.Keys {
					params[k] = chi.RouteContext(r.Context()).URLParams.Values[i]
				}

				status := ww.Status()
				method := r.Method
				host := r.Host
				route := chi.RouteContext(r.Context()).RoutePattern()
				end := time.Now()
				latency := end.Sub(start)
				userAgent := r.UserAgent()
				ip := r.RemoteAddr
				referer := r.Referer()

				baseAttributes := []slog.Attr{}

				requestAttributes := []slog.Attr{
					slog.Time("time", start),
					slog.String("method", method),
					slog.String("host", host),
					slog.String("path", path),
					slog.String("query", query),
					slog.Any("params", params),
					slog.String("route", route),
					slog.String("ip", ip),
					slog.String("referer", referer),
				}

				responseAttributes := []slog.Attr{
					slog.Time("time", end),
					slog.Duration("latency", latency),
					slog.Int("status", status),
				}

				if config.WithRequestID {
					baseAttributes = append(baseAttributes, slog.String(RequestIDKey, middleware.GetReqID(r.Context())))
				}

				// otel
				baseAttributes = append(baseAttributes, extractTraceSpanID(r.Context(), config.WithTraceID, config.WithSpanID)...)

				// request body
				requestAttributes = append(requestAttributes, slog.Int("length", br.bytes))
				if config.WithRequestBody {
					requestAttributes = append(requestAttributes, slog.String("body", br.body.String()))
				}

				// request headers
				if config.WithRequestHeader {
					kv := []any{}

					for k, v := range r.Header {
						if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					requestAttributes = append(requestAttributes, slog.Group("header", kv...))
				}

				if config.WithUserAgent {
					requestAttributes = append(requestAttributes, slog.String("user-agent", userAgent))
				}

				// response body
				responseAttributes = append(responseAttributes, slog.Int("length", ww.BytesWritten()))
				if config.WithResponseBody {
					responseAttributes = append(responseAttributes, slog.String("body", bw.body.String()))
				}

				// response headers
				if config.WithResponseHeader {
					kv := []any{}

					for k, v := range w.Header() {
						if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					responseAttributes = append(responseAttributes, slog.Group("header", kv...))
				}

				attributes := append(
					[]slog.Attr{
						{
							Key:   "request",
							Value: slog.GroupValue(requestAttributes...),
						},
						{
							Key:   "response",
							Value: slog.GroupValue(responseAttributes...),
						},
					},
					baseAttributes...,
				)

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

func extractTraceSpanID(ctx context.Context, withTraceID bool, withSpanID bool) []slog.Attr {
	if !(withTraceID || withSpanID) {
		return []slog.Attr{}
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return []slog.Attr{}
	}

	attrs := []slog.Attr{}
	spanCtx := span.SpanContext()

	if withTraceID && spanCtx.HasTraceID() {
		traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
		attrs = append(attrs, slog.String(TraceIDKey, traceID))
	}

	if withSpanID && spanCtx.HasSpanID() {
		spanID := spanCtx.SpanID().String()
		attrs = append(attrs, slog.String(SpanIDKey, spanID))
	}

	return attrs
}
