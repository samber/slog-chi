package slogchi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHostFilters(t *testing.T) {
	tests := []struct {
		name     string
		accept   Filter
		ignore   Filter
		matching string
		other    string
	}{
		{"exact", AcceptHost("api.example.com"), IgnoreHost("api.example.com"), "api.example.com", "other.example.com"},
		{"exact_port", AcceptHost("api.example.com:8443"), IgnoreHost("api.example.com:8443"), "api.example.com:8443", "api.example.com:9443"},
		{"contains", AcceptHostContains("example.com"), IgnoreHostContains("example.com"), "api.example.com:8443", "api.other.test:8443"},
		{"prefix", AcceptHostPrefix("api."), IgnoreHostPrefix("api."), "api.example.com:8443", "www.example.com:8443"},
		{"suffix", AcceptHostSuffix(".example.com"), IgnoreHostSuffix(".example.com"), "api.example.com", "api.example.com:8443"},
		{"suffix_port", AcceptHostSuffix(":8443"), IgnoreHostSuffix(":8443"), "api.example.com:8443", "api.example.com:9443"},
		{"regexp", AcceptHostMatch(*regexp.MustCompile(`^api\.example\.com:[0-9]+$`)), IgnoreHostMatch(*regexp.MustCompile(`^api\.example\.com:[0-9]+$`)), "api.example.com:8443", "www.example.com:8443"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := []struct {
				name    string
				host    string
				urlHost string
				match   bool
			}{
				{"incoming_match", test.matching, "", true},
				{"incoming_other", test.other, "", false},
				{"host_overrides_url_match", test.matching, test.other, true},
				{"host_overrides_url_other", test.other, test.matching, false},
			}

			for _, request := range requests {
				t.Run(request.name, func(t *testing.T) {
					r := httptest.NewRequest(http.MethodGet, "/", nil)
					r.Host = request.host
					r.URL.Host = request.urlHost
					if got := test.accept(nil, r); got != request.match {
						t.Errorf("accept = %v, want %v for Host %q and URL.Host %q", got, request.match, r.Host, r.URL.Host)
					}
					if got := test.ignore(nil, r); got != !request.match {
						t.Errorf("ignore = %v, want %v for Host %q and URL.Host %q", got, !request.match, r.Host, r.URL.Host)
					}
				})
			}
		})
	}
}

func TestHostFiltersWithServer(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		host   string
		logged bool
	}{
		{"accept_match", AcceptHost("api.example.com:8443"), "api.example.com:8443", true},
		{"accept_other", AcceptHost("api.example.com:8443"), "other.example.com:8443", false},
		{"ignore_match", IgnoreHost("api.example.com:8443"), "api.example.com:8443", false},
		{"ignore_other", IgnoreHost("api.example.com:8443"), "other.example.com:8443", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			done := make(chan struct{})
			router := chi.NewRouter()
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer close(done)
					next.ServeHTTP(w, r)
				})
			})
			router.Use(NewWithFilters(slog.New(slog.NewJSONHandler(&output, nil)), test.filter))
			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Host != "" {
					t.Errorf("incoming URL.Host = %q, want empty", r.URL.Host)
				}
				w.Header().Set("X-Test", "kept")
				w.WriteHeader(http.StatusAccepted)
				_, _ = io.WriteString(w, "response")
			})
			server := httptest.NewServer(router)
			defer server.Close()

			request, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Host = test.host
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Error(err)
				}
			}()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			<-done

			if response.StatusCode != http.StatusAccepted || string(body) != "response" || response.Header.Get("X-Test") != "kept" {
				t.Errorf("response = %d %q, X-Test %q", response.StatusCode, body, response.Header.Get("X-Test"))
			}
			if !test.logged {
				if output.Len() != 0 {
					t.Errorf("unexpected log: %s", &output)
				}
				return
			}

			var record struct {
				Request struct {
					Host string
				}
				Response struct {
					Status int
				}
			}
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatalf("decode log %q: %v", output.String(), err)
			}
			if record.Request.Host != test.host || record.Response.Status != http.StatusAccepted {
				t.Errorf("logged host = %q, status = %d", record.Request.Host, record.Response.Status)
			}
		})
	}
}
