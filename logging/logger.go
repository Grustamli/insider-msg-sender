package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"time"
)

type Level string

type LogConfig struct {
	IsProduction bool
	Level        Level
}

const (
	TRACE Level = "TRACE"
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
	PANIC Level = "PANIC"
)

func New(cfg LogConfig) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if cfg.IsProduction {
		return prodLogger(cfg.Level)
	}
	return devLogger(cfg.Level)
}

func prodLogger(level Level) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(logLevelToZero(level)).
		With().
		Timestamp().
		Logger()
}

func devLogger(level Level) zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
	})).
		Level(logLevelToZero(level)).
		With().
		Timestamp().
		Logger()
}

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
