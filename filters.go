package slogchi

import (
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

type Filter func(ww middleware.WrapResponseWriter, r *http.Request) bool

// Basic
func Accept(filter Filter) Filter { return filter }
func Ignore(filter Filter) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool { return !filter(ww, r) }
}

// Method
func AcceptMethod(methods ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return true
			}
		}

		return false
	}
}

func IgnoreMethod(methods ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return false
			}
		}

		return true
	}
}

// Status
func AcceptStatus(statuses ...int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(statuses, ww.Status())
	}
}

func IgnoreStatus(statuses ...int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(statuses, ww.Status())
	}
}

func AcceptStatusGreaterThan(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() > status
	}
}

func AcceptStatusGreaterThanOrEqual(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() >= status
	}
}

func AcceptStatusLessThan(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() < status
	}
}

func AcceptStatusLessThanOrEqual(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() <= status
	}
}

func IgnoreStatusGreaterThan(status int) Filter {
	return AcceptStatusLessThanOrEqual(status)
}

func IgnoreStatusGreaterThanOrEqual(status int) Filter {
	return AcceptStatusLessThan(status)
}

func IgnoreStatusLessThan(status int) Filter {
	return AcceptStatusGreaterThanOrEqual(status)
}

func IgnoreStatusLessThanOrEqual(status int) Filter {
	return AcceptStatusGreaterThan(status)
}

// Path
func AcceptPath(urls ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(urls, r.URL.Path)
	}
}

func IgnorePath(urls ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(urls, r.URL.Path)
	}
}

func AcceptPathContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return true
			}
		}

		return false
	}
}

func IgnorePathContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return false
			}
		}

		return true
	}
}

func AcceptPathPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathSuffix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathSuffix(suffixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(r.URL.Path, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return true
			}
		}

		return false
	}
}

func IgnorePathMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return false
			}
		}

		return true
	}
}

// Host
func AcceptHost(hosts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(hosts, r.URL.Host)
	}
}

func IgnoreHost(hosts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(hosts, r.URL.Host)
	}
}

func AcceptHostContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Host, part) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Host, part) {
				return false
			}
		}

		return true
	}
}

func AcceptHostPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostSuffix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostSuffix(suffixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(r.URL.Host, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return false
			}
		}

		return true
	}
}
