package slogchi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type customAttributesCtxKeyType struct{}

var customAttributesCtxKey = customAttributesCtxKeyType{}

var (
	TraceIDKey   = "trace_id"
	SpanIDKey    = "span_id"
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
	WithClientIP       bool
	WithCustomMessage  func(w http.ResponseWriter, r *http.Request) string

	Filters []Filter
}

// New returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func New(logger *slog.Logger) func(http.Handler) http.Handler {
	return NewWithConfig(logger, DefaultConfig())
}

// NewWithFilters returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewWithFilters(logger *slog.Logger, filters ...Filter) func(http.Handler) http.Handler {
	config := DefaultConfig()
	config.Filters = filters
	return NewWithConfig(logger, config)
}

// DefaultConfig returns the default configuration for the request logger.
func DefaultConfig() Config {
	return Config{
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
		WithClientIP:       true,
		WithCustomMessage:  nil,

		Filters: []Filter{},
	}
}

// NewWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			query := r.URL.RawQuery

			// Make sure we create a map only once per request (in case we have multiple middleware instances)
			if v := r.Context().Value(customAttributesCtxKey); v == nil {
				r = r.WithContext(context.WithValue(r.Context(), customAttributesCtxKey, &sync.Map{}))
			}

			// Apply filters BEFORE reading the request body to avoid unnecessary memory allocation.
			// This is important when using filters with WithRequestBody=true, as the body
			// capture can consume significant memory for large requests.
			filteredOut := false
			if len(config.Filters) > 0 {
				filteredOut = filterEarly(config.Filters, w, r)
			}

			// dump request body only if not filtered out
			var br *bodyReader
			if !filteredOut {
				br = newBodyReader(r.Body, RequestBodyMaxSize, config.WithRequestBody)
				r.Body = br
			}

			// dump response body
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var bw *bodyWriter
			if config.WithResponseBody && !filteredOut {
				bw = newBodyWriter(ResponseBodyMaxSize)
				ww.Tee(bw)
			}

			defer func() {
				// Apply filters after response is written (for status-based filters)
				if !filteredOut && len(config.Filters) > 0 {
					for _, filter := range config.Filters {
						if !filter(ww, r) {
							filteredOut = true
							break
						}
					}
				}

				if filteredOut {
					return
				}

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

				// Get the body reader if it was created
				var body string
				var bodyLength int
				if br != nil {
					bodyLength = br.bytes
					body = br.body.String()
				}

				baseAttributes := make([]slog.Attr, 0, 3)
				requestAttributes := make([]slog.Attr, 0, 13)
				responseAttributes := make([]slog.Attr, 0, 6)

				requestAttributes = append(requestAttributes,
					slog.Time("time", start.UTC()),
					slog.String("method", method),
					slog.String("host", host),
					slog.String("path", path),
					slog.String("query", query),
					slog.Any("params", params),
					slog.String("route", route),
					slog.String("referer", referer),
				)

				if config.WithClientIP {
					requestAttributes = append(requestAttributes, slog.String("ip", ip))
				}

				responseAttributes = append(responseAttributes,
					slog.Time("time", end.UTC()),
					slog.Duration("latency", latency),
					slog.Int("status", status),
				)

				if config.WithRequestID {
					baseAttributes = append(baseAttributes, slog.String(RequestIDKey, middleware.GetReqID(r.Context())))
				}

				// otel
				baseAttributes = append(baseAttributes, extractTraceSpanID(r.Context(), config.WithTraceID, config.WithSpanID)...)

				// request body
				requestAttributes = append(requestAttributes, slog.Int("length", bodyLength))
				if config.WithRequestBody {
					requestAttributes = append(requestAttributes, slog.String("body", body))
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
				if config.WithResponseBody && bw != nil {
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
					if m, ok := v.(*sync.Map); ok {
						m.Range(func(key, value any) bool {
							attributes = append(attributes, slog.Attr{Key: key.(string), Value: value.(slog.Value)})
							return true
						})
					}
				}

				level := config.DefaultLevel
				if status >= http.StatusInternalServerError {
					level = config.ServerErrorLevel
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = config.ClientErrorLevel
				}

				msg := strconv.Itoa(status) + ": " + http.StatusText(status)
				if config.WithCustomMessage != nil {
					msg = config.WithCustomMessage(ww, r)
				}

				logger.LogAttrs(r.Context(), level, msg, attributes...)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// filterEarly checks filters before the request is processed.
// Returns true if the request should be filtered out (not logged).
func filterEarly(filters []Filter, w http.ResponseWriter, r *http.Request) bool {
	// Create a preview writer to check filters
	pw := &previewWriter{ResponseWriter: w}

	// Check all filters
	for _, filter := range filters {
		if !filter(pw, r) {
			return true
		}
	}

	return false
}

// previewWriter wraps http.ResponseWriter to capture status for early filter checks
type previewWriter struct {
	http.ResponseWriter
	status int
}

func (p *previewWriter) WriteHeader(status int) {
	p.status = status
	p.ResponseWriter.WriteHeader(status)
}

func (p *previewWriter) Status() int {
	return p.status
}

func (p *previewWriter) BytesWritten() int {
	return 0
}

func (p *previewWriter) Unwrap() http.ResponseWriter {
	return p.ResponseWriter
}

func (p *previewWriter) Discard() {}

func (p *previewWriter) Tee(w io.Writer) {
	// Not used in preview mode
}

// statusFilterKey is the context key for storing status-based filters
type statusFilterKey struct{}

var statusFilterKeyInstance = statusFilterKey{}

// previewWriterKey is the context key for storing the preview writer
type previewWriterKey struct{}

var previewWriterKeyInstance = previewWriterKey{}

// AddCustomAttributes adds custom attributes to the request context. This func can be called from any handler or middleware, as long as the slog-chi middleware is already mounted.
func AddCustomAttributes(r *http.Request, attrs ...slog.Attr) {
	AddContextAttributes(r.Context(), attrs...)
}

// AddContextAttributes is the same as AddCustomAttributes, but it doesn't need access to the request struct.
func AddContextAttributes(ctx context.Context, attrs ...slog.Attr) {
	if v := ctx.Value(customAttributesCtxKey); v != nil {
		if m, ok := v.(*sync.Map); ok {
			for _, attr := range attrs {
				m.Store(attr.Key, attr.Value)
			}
		}
	}
}

func extractTraceSpanID(ctx context.Context, withTraceID bool, withSpanID bool) []slog.Attr {
	if !withTraceID && !withSpanID {
		return []slog.Attr{}
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return []slog.Attr{}
	}

	attrs := make([]slog.Attr, 0, 2)
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
