// Package config standardizes config for our application.
// This package must not be imported anywhere except `main` packages.
// Anything that requires the config deeper than the main package must
// be configured with dependency injection, passing config down through
// the intermediate layers.
package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	// init will run automatically when the config package is imported.

	// Automatically merge env variables into any config files
	// or CLI vars.
	// Viper will take the dot path of any variable we try and
	// lookup and as the final fallback, look in the environment
	// by converting hypens to underscores and uppercasing.
	// e.g. "database.host" will be looked up as "DATABASE_HOST"
	viper.AutomaticEnv()

	// Set the config file
	viper.SetConfigFile("config.toml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	viper.SetDefault("log.level", "info")
	log_level := viper.GetString("log.level")
	level, err := logrus.ParseLevel(log_level)
	if err != nil {
		panic(fmt.Errorf("fatal parsing log level: %w", err))
	}
	logrus.SetLevel(level)
}
