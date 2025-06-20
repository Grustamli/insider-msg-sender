// Package logging provides utilities for configuring and initializing zerolog loggers
// with sensible defaults for development and production environments.
package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Level represents the logging severity level.
type Level string

// LogConfig holds options for constructing a zerolog.Logger.
// IsProduction toggles between production and development styles.
// Level sets the minimum log level to emit.
type LogConfig struct {
	IsProduction bool  // true for production settings, false for development
	Level        Level // minimum log level (TRACE, DEBUG, INFO, WARN, ERROR, PANIC)
}

// Supported logging levels.
const (
	TRACE Level = "TRACE" // Trace level logs are highly detailed and typically only enabled during development.
	DEBUG Level = "DEBUG" // Debug level logs provide diagnostic information useful for debugging.
	INFO  Level = "INFO"  // Info level logs convey general operational entries about application progress.
	WARN  Level = "WARN"  // Warn level logs indicate potentially harmful situations.
	ERROR Level = "ERROR" // Error level logs indicate error events that might still allow the application to continue.
	PANIC Level = "PANIC" // Panic level logs indicate very severe error events that lead to a program panic.
)

// New creates and returns a configured zerolog.Logger based on the provided LogConfig.
// It sets timestamp formatting to Unix milliseconds and attaches stack trace marshaling for errors.
// For production, it returns a JSON logger; for development, a human-friendly console logger.
func New(cfg LogConfig) zerolog.Logger {
	// use Unix ms timestamps for all loggers
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	// enable pkg/errors stack trace marshaling
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if cfg.IsProduction {
		return prodLogger(cfg.Level)
	}
	return devLogger(cfg.Level)
}

// prodLogger returns a zerolog.Logger that writes JSON-formatted logs to stdout
// at the specified log level, including timestamps.
func prodLogger(level Level) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(logLevelToZero(level)).
		With().
		Timestamp().
		Logger()
}

// devLogger returns a zerolog.Logger that writes human-readable console logs to stdout
// at the specified log level, including RFC3339 timestamps.
func devLogger(level Level) zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
	})).
		Level(logLevelToZero(level)).
		With().
		Timestamp().
		Logger()
}

// logLevelToZero maps our Level type to zerolog.Level constants.
// If the provided level is unrecognized, INFO is used as the default.
func logLevelToZero(level Level) zerolog.Level {
	switch level {
	case PANIC:
		return zerolog.PanicLevel
	case ERROR:
		return zerolog.ErrorLevel
	case WARN:
		return zerolog.WarnLevel
	case INFO:
		return zerolog.InfoLevel
	case DEBUG:
		return zerolog.DebugLevel
	case TRACE:
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}
