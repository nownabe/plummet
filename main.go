package main

import (
	"github.com/nownabe/plummet/pkg/chibi"
	"github.com/nownabe/plummet/pkg/slack"
)

func main() {
	cfg, err := initConfig()
	if err != nil {
		panic(err)
	}

	if err := chibi.InitLogger(cfg.LogLevel, cfg.LogPretty); err != nil {
		panic(err)
	}

	av := newAlphaVantage(cfg.AlphaVantageAPIKey)
	s := slack.New(cfg.SlackToken)

	ap := &app{
		symbols:      cfg.Symbols,
		slackChannel: cfg.SlackChannel,
		alphaVantage: av,
		slack:        s,
	}

	chibi.Serve(ap.handler(), cfg.Port)
}
