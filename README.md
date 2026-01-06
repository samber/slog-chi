
# slog: Chi middleware

[![tag](https://img.shields.io/github/tag/samber/slog-chi.svg)](https://github.com/samber/slog-chi/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/slog-chi?status.svg)](https://pkg.go.dev/github.com/samber/slog-chi)
![Build Status](https://github.com/samber/slog-chi/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/slog-chi)](https://goreportcard.com/report/github.com/samber/slog-chi)
[![Coverage](https://img.shields.io/codecov/c/github/samber/slog-chi)](https://codecov.io/gh/samber/slog-chi)
[![Contributors](https://img.shields.io/github/contributors/samber/slog-chi)](https://github.com/samber/slog-chi/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/slog-chi)](./LICENSE)

[Chi](https://github.com/go-chi/chi) middleware to log http requests using [slog](https://pkg.go.dev/log/slog).

<div align="center">
  <hr>
  <sup><b>Sponsored by:</b></sup>
  <br>
  <a href="https://www.dash0.com?utm_campaign=148395251-samber%20github%20sponsorship&utm_source=github&utm_medium=sponsorship&utm_content=samber">
    <div>
      <img src="https://github.com/user-attachments/assets/b1f2e876-0954-4dc3-824d-935d29ba8f3f" width="200" alt="Dash0">
    </div>
    <div>
      100% OpenTelemetry-native observability platform<br>Simple to use, built on open standards, and designed for full cost control
    </div>
  </a>
  <hr>
</div>

**See also:**

- [slog-multi](https://github.com/samber/slog-multi): `slog.Handler` chaining, fanout, routing, failover, load balancing...
- [slog-formatter](https://github.com/samber/slog-formatter): `slog` attribute formatting
- [slog-sampling](https://github.com/samber/slog-sampling): `slog` sampling policy
- [slog-mock](https://github.com/samber/slog-mock): `slog.Handler` for test purposes

**HTTP middlewares:**

- [slog-gin](https://github.com/samber/slog-gin): Gin middleware for `slog` logger
- [slog-echo](https://github.com/samber/slog-echo): Echo middleware for `slog` logger
- [slog-fiber](https://github.com/samber/slog-fiber): Fiber middleware for `slog` logger
- [slog-chi](https://github.com/samber/slog-chi): Chi middleware for `slog` logger
- [slog-http](https://github.com/samber/slog-http): `net/http` middleware for `slog` logger

**Loggers:**

- [slog-zap](https://github.com/samber/slog-zap): A `slog` handler for `Zap`
- [slog-zerolog](https://github.com/samber/slog-zerolog): A `slog` handler for `Zerolog`
- [slog-logrus](https://github.com/samber/slog-logrus): A `slog` handler for `Logrus`

**Log sinks:**

- [slog-datadog](https://github.com/samber/slog-datadog): A `slog` handler for `Datadog`
- [slog-betterstack](https://github.com/samber/slog-betterstack): A `slog` handler for `Betterstack`
- [slog-rollbar](https://github.com/samber/slog-rollbar): A `slog` handler for `Rollbar`
- [slog-loki](https://github.com/samber/slog-loki): A `slog` handler for `Loki`
- [slog-sentry](https://github.com/samber/slog-sentry): A `slog` handler for `Sentry`
- [slog-syslog](https://github.com/samber/slog-syslog): A `slog` handler for `Syslog`
- [slog-logstash](https://github.com/samber/slog-logstash): A `slog` handler for `Logstash`
- [slog-fluentd](https://github.com/samber/slog-fluentd): A `slog` handler for `Fluentd`
- [slog-graylog](https://github.com/samber/slog-graylog): A `slog` handler for `Graylog`
- [slog-quickwit](https://github.com/samber/slog-quickwit): A `slog` handler for `Quickwit`
- [slog-slack](https://github.com/samber/slog-slack): A `slog` handler for `Slack`
- [slog-telegram](https://github.com/samber/slog-telegram): A `slog` handler for `Telegram`
- [slog-mattermost](https://github.com/samber/slog-mattermost): A `slog` handler for `Mattermost`
- [slog-microsoft-teams](https://github.com/samber/slog-microsoft-teams): A `slog` handler for `Microsoft Teams`
- [slog-webhook](https://github.com/samber/slog-webhook): A `slog` handler for `Webhook`
- [slog-kafka](https://github.com/samber/slog-kafka): A `slog` handler for `Kafka`
- [slog-nats](https://github.com/samber/slog-nats): A `slog` handler for `NATS`
- [slog-parquet](https://github.com/samber/slog-parquet): A `slog` handler for `Parquet` + `Object Storage`
- [slog-channel](https://github.com/samber/slog-channel): A `slog` handler for Go channels

## 🚀 Install

```sh
go get github.com/samber/slog-chi
```

**Compatibility**: go >= 1.21

No breaking changes will be made to exported APIs before v2.0.0.

## 💡 Usage

### Handler options

```go
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

	Filters []Filter
}
```

Attributes will be injected in log payload.

Other global parameters:

```go
slogchi.TraceIDKey = "trace_id"
slogchi.SpanIDKey = "span_id"
slogchi.RequestIDKey = "id"
slogchi.RequestBodyMaxSize  = 64 * 1024 // 64KB
slogchi.ResponseBodyMaxSize = 64 * 1024 // 64KB
slogchi.HiddenRequestHeaders = map[string]struct{}{ ... }
slogchi.HiddenResponseHeaders = map[string]struct{}{ ... }
```

### Minimal

```go
import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
)

// Create a slog logger, which:
//   - Logs to stdout.
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))
router.Use(middleware.Recoverer)

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})
router.GET("/error", func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(400), 400)
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg="200: OK" env=production request.time=2023-10-15T20:32:58.626+02:00 request.method=GET request.path=/ request.route="" request.ip=127.0.0.1:63932 request.length=0 response.time=2023-10-15T20:32:58.926+02:00 response.latency=100ms response.status=200 response.length=7 id=""
```

### OTEL

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

config := slogchi.Config{
	WithSpanID:  true,
	WithTraceID: true,
}

router := chi.NewRouter()
router.Use(slogchi.NewWithConfig(logger, config))
router.Use(middleware.Recoverer)
```

### Custom log levels

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

config := slogchi.Config{
	DefaultLevel:     slog.LevelInfo,
	ClientErrorLevel: slog.LevelWarn,
	ServerErrorLevel: slog.LevelError,
}

router := chi.NewRouter()
router.Use(slogchi.NewWithConfig(logger, config))
router.Use(middleware.Recoverer)
```

### Verbose

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

config := slogchi.Config{
	WithRequestBody: true,
	WithResponseBody: true,
	WithRequestHeader: true,
	WithResponseHeader: true,
}

router := chi.NewRouter()
router.Use(slogchi.NewWithConfig(logger, config))
router.Use(middleware.Recoverer)
```

### Filters

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

router := chi.NewRouter()
router.Use(
	slogchi.NewWithFilters(
		logger,
		slogchi.Accept(func (ww middleware.WrapResponseWriter, r *http.Request) bool {
			return xxx
		}),
		slogchi.IgnoreStatus(401, 404),
	),
)
router.Use(middleware.Recoverer)
```

Available filters:
- Accept / Ignore
- AcceptMethod / IgnoreMethod
- AcceptStatus / IgnoreStatus
- AcceptStatusGreaterThan / IgnoreStatusGreaterThan
- AcceptStatusLessThan / IgnoreStatusLessThan
- AcceptStatusGreaterThanOrEqual / IgnoreStatusGreaterThanOrEqual
- AcceptStatusLessThanOrEqual / IgnoreStatusLessThanOrEqual
- AcceptPath / IgnorePath
- AcceptPathContains / IgnorePathContains
- AcceptPathPrefix / IgnorePathPrefix
- AcceptPathSuffix / IgnorePathSuffix
- AcceptPathMatch / IgnorePathMatch
- AcceptHost / IgnoreHost
- AcceptHostContains / IgnoreHostContains
- AcceptHostPrefix / IgnoreHostPrefix
- AcceptHostSuffix / IgnoreHostSuffix
- AcceptHostMatch / IgnoreHostMatch

### Using custom time formatters

```go
import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
)

// Create a slog logger, which:
//   - Logs to stdout.
//   - RFC3339 with UTC time format.
logger := slog.New(
	slogformatter.NewFormatterHandler(
		slogformatter.TimezoneConverter(time.UTC),
		slogformatter.TimeFormatter(time.DateTime, nil),
	)(
		slog.NewTextHandler(os.Stdout, nil),
	),
)

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))
router.Use(middleware.Recoverer)

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})
router.GET("/error", func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(400), 400)
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg="200: OK" env=production request.time=2023-10-15T20:32:58.626Z request.method=GET request.path=/ request.route="" request.ip=127.0.0.1:63932 request.length=0 response.time=2023-10-15T20:32:58Z response.latency=100ms response.status=200 response.length=7 id=""
```

### Using custom logger sub-group

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger.WithGroup("http")))
router.Use(middleware.Recoverer)

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})
router.GET("/error", func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(400), 400)
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg="200: OK" env=production http.request.time=2023-10-15T20:32:58.626+02:00 http.request.method=GET http.request.path=/ http.request.route="" http.request.ip=127.0.0.1:63932 http.request.length=0 http.response.time=2023-10-15T20:32:58.926+02:00 http.response.latency=100ms http.response.status=200 http.response.length=7 http.id=""
```

### Adding custom attributes

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Add an attribute to all log entries made through this logger.
logger = logger.With("env", "production")

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))
router.Use(middleware.Recoverer)

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	// Add an attribute to a single log entry.
	slogchi.AddCustomAttributes(r, slog.String("foo", "bar"))
	w.Write([]byte("Hello, World!"))
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg="200: OK" env=production request.time=2023-10-15T20:32:58.626+02:00 request.method=GET request.path=/ request.route="" request.ip=127.0.0.1:63932 request.length=0 response.time=2023-10-15T20:32:58.926+02:00 response.latency=100ms response.status=200 response.length=7 id="" foo=bar
```

### JSON output

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))
router.Use(middleware.Recoverer)

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// {"time":"2023-10-15T20:32:58.926+02:00","level":"INFO","msg":"200: OK","env":"production","http":{"request":{"time":"2023-10-15T20:32:58.626+02:00","method":"GET","path":"/","route":"","ip":"127.0.0.1:55296","length":0},"response":{"time":"2023-10-15T20:32:58.926+02:00","latency":100000,"status":200,"length":7},"id":""}}
```

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/slog-chi)
- Fix [open issues](https://github.com/samber/slog-chi/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/slog-chi)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2023 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
