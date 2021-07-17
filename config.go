package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	AlphaVantageAPIKey string `required:"true" split_words:"true"`
	LogLevel           string `default:"info" split_words:"true"`
	LogPretty          bool   `default:"false" split_words:"true"`
	Port               string `default:"8080" envconfig:"PORT"`
	SlackToken         string `required:"true" split_words:"true"`
	SlackChannel       string `required:"true" split_words:"true"`
	Symbols            []string
}

func initConfig() (*config, error) {
	c := &config{}

	if err := envconfig.Process("plummet", c); err != nil {
		return nil, err
	}

	return c, nil
}
