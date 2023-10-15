package main

import (
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
	slogformatter "github.com/samber/slog-formatter"
)

func main() {
	// Create a slog logger, which:
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	logger := slog.New(
		slogformatter.NewFormatterHandler(
			slogformatter.TimezoneConverter(time.UTC),
			slogformatter.TimeFormatter(time.RFC3339, nil),
		)(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
		),
	)

	// Add an attribute to all log entries made through this logger.
	logger = logger.With("env", "production")

	// Chi instance
	r := chi.NewRouter()

	// Middleware
	// config := slogchi.Config{WithRequestBody: true, WithResponseBody: true, WithRequestHeader: true, WithResponseHeader: true}
	// r.Use(slogchi.NewWithConfig(logger, config))
	r.Use(slogchi.New(logger.WithGroup("http")))

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/foobar/*", func(w http.ResponseWriter, r *http.Request) {
		slogchi.AddCustomAttributes(r, slog.String("foo", "bar"))
		w.Write([]byte("welcome"))
	})
	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(400), 400)
	})

	// Start server
	err := http.ListenAndServe(":4242", r)
	if err != nil {
		logger.Error(err.Error())
	}

	// output:
	// time=2023-10-15T20:32:58.926+02:00 level=INFO msg=OK env=production http.time=2023-10-15T18:32:58Z http.latency=20.834µs http.method=GET http.path=/ http.status=200 http.user-agent=curl/7.77.0
}
