package observability

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configures the global zerolog logger based on level + format.
// format: "json" | "console". level: "debug" | "info" | "warn" | "error".
func Setup(level, format, env, service string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	var w io.Writer = os.Stdout
	if format == "console" {
		w = zerolog.ConsoleWriter{Out: os.Stderr}
	}

	log.Logger = zerolog.New(w).
		With().
		Timestamp().
		Str("service", service).
		Str("env", env).
		Logger()
}
