module example

go 1.21

replace github.com/samber/slog-chi => ../

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/samber/slog-chi v1.0.0
	github.com/samber/slog-formatter v1.0.0
)

require (
	github.com/samber/lo v1.38.1 // indirect
	github.com/samber/slog-multi v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
)
