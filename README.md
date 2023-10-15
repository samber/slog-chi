
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

**See also:**

- [slog-multi](https://github.com/samber/slog-multi): `slog.Handler` chaining, fanout, routing, failover, load balancing...
- [slog-formatter](https://github.com/samber/slog-formatter): `slog` attribute formatting
- [slog-sampling](https://github.com/samber/slog-sampling): `slog` sampling policy
- [slog-gin](https://github.com/samber/slog-gin): Gin middleware for `slog` logger
- [slog-echo](https://github.com/samber/slog-echo): Echo middleware for `slog` logger
- [slog-fiber](https://github.com/samber/slog-fiber): Fiber middleware for `slog` logger
- [slog-datadog](https://github.com/samber/slog-datadog): A `slog` handler for `Datadog`
- [slog-rollbar](https://github.com/samber/slog-rollbar): A `slog` handler for `Rollbar`
- [slog-sentry](https://github.com/samber/slog-sentry): A `slog` handler for `Sentry`
- [slog-syslog](https://github.com/samber/slog-syslog): A `slog` handler for `Syslog`
- [slog-logstash](https://github.com/samber/slog-logstash): A `slog` handler for `Logstash`
- [slog-fluentd](https://github.com/samber/slog-fluentd): A `slog` handler for `Fluentd`
- [slog-graylog](https://github.com/samber/slog-graylog): A `slog` handler for `Graylog`
- [slog-loki](https://github.com/samber/slog-loki): A `slog` handler for `Loki`
- [slog-slack](https://github.com/samber/slog-slack): A `slog` handler for `Slack`
- [slog-telegram](https://github.com/samber/slog-telegram): A `slog` handler for `Telegram`
- [slog-mattermost](https://github.com/samber/slog-mattermost): A `slog` handler for `Mattermost`
- [slog-microsoft-teams](https://github.com/samber/slog-microsoft-teams): A `slog` handler for `Microsoft Teams`
- [slog-webhook](https://github.com/samber/slog-webhook): A `slog` handler for `Webhook`
- [slog-kafka](https://github.com/samber/slog-kafka): A `slog` handler for `Kafka`
- [slog-parquet](https://github.com/samber/slog-parquet): A `slog` handler for `Parquet` + `Object Storage`
- [slog-zap](https://github.com/samber/slog-zap): A `slog` handler for `Zap`
- [slog-zerolog](https://github.com/samber/slog-zerolog): A `slog` handler for `Zerolog`
- [slog-logrus](https://github.com/samber/slog-logrus): A `slog` handler for `Logrus`

## 🚀 Install

```sh
go get github.com/samber/slog-chi
```

**Compatibility**: go >= 1.21

No breaking changes will be made to exported APIs before v2.0.0.

## 💡 Usage

### Minimal

```go
import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
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
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg=OK env=production http.time=2023-10-15T18:32:58Z http.latency=20.834µs http.method=GET http.path=/ http.status=200 http.user-agent=curl/7.77.0
```

### Filters

```go
import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
)

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
```

Available filters:
- Accept / Ignore
- AcceptMethod / IgnoreMethod
- AcceptStatus / IgnoreStatus
- AcceptStatusGreaterThan / IgnoreStatusLessThan
- AcceptStatusGreaterThanOrEqual / IgnoreStatusLessThanOrEqual
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
// time=2023-04-10T14:00:00Z level=INFO msg="Success"  status=200 method=GET path=/ ip=::1 latency=25.958µs user-agent=curl/7.77.0 time=2023-04-10T14:00:00Z request-id=229c7fc8-64f5-4467-bc4a-940700503b0d
```

### Using custom logger sub-group

```go
import (
	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
)

// Create a slog logger, which:
//   - Logs to stdout.
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger.WithGroup("http")))

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
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg=OK env=production http.time=2023-10-15T18:32:58Z http.latency=20.834µs http.method=GET http.path=/ http.status=200 http.user-agent=curl/7.77.0
```

### Adding custom attributes

```go
import (
	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
)

// Create a slog logger, which:
//   - Logs to stdout.
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Add an attribute to all log entries made through this logger.
logger = logger.With("env", "production")

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// time=2023-10-15T20:32:58.926+02:00 level=INFO msg=OK env=production http.time=2023-10-15T18:32:58Z http.latency=20.834µs http.method=GET http.path=/ http.status=200 http.user-agent=curl/7.77.0
```

### JSON output

```go
import (
	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
)

// Create a slog logger, which:
//   - Logs to stdout.
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Chi instance
router := chi.NewRouter()

// Middleware
router.Use(slogchi.New(logger))

// Routes
router.GET("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
})

// Start server
err := http.ListenAndServe(":4242", router)

// output:
// {"time":"2023-10-15T20:48:12.859856+02:00","level":"INFO","msg":"OK","env":"production","time":"2023-10-15T18:48:12Z","latency":4417,"method":"GET","path":"/","status":200,"user-agent":"curl/7.77.0"}
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
