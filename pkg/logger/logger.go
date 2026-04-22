package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", "payment-gateway").
			Logger()
	}

	return log.Logger
}

func Get() zerolog.Logger {
	return log.Logger
}
