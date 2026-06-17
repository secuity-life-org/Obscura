package main

// Build info, injected via -ldflags at build time. Defaults are used for
// `go run` / unstamped builds.
var (
	version   = "9.0.0"
	commit    = "dev"
	buildDate = "unknown"
)
