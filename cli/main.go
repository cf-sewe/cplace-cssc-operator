package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cf-sewe/cplace-cssc-operator/internal/config"
	"github.com/cf-sewe/cplace-cssc-operator/internal/operator"
	"github.com/joho/godotenv"
)

// main initializes the environment and starts the workers and routes
func main() {
	// configure zerolog ConsoleWriter
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Caller().Logger()
	//log.Logger = log.Logger.With().Caller().Logger()
	log.Info().Msg("cplace Cloud Operator is starting")

	// Load config from .env file
	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Unable to load .env file")
	}

	// This code sets the log level for the operator, defaulting to WARN.
	// The log level can be set via the environment variable CSSC_OPERATOR_LOG_LEVEL.
	// The log level can be one of the following values: debug, info, warn, error, fatal, panic.
	// If the environment variable is not set or is set to an invalid value, the default value will be used.
	var logLevel = zerolog.WarnLevel
	ll, err := zerolog.ParseLevel(config.GetEnv(config.LogLevel, "warn"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid log level, using default")
	} else {
		logLevel = ll
	}
	log.Logger.Level(logLevel)

	ctx := context.Background()

	// Initialize the Swarm environment with the given context
	op, err := operator.NewOperator(&ctx, &log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize cplace Cloud Operator")
	}
	// do something with env
	op.Swarm.Bogus()

	log.Info().Msg("cplace Cloud Operator has stopped")
}
