package main

import (
	"time"

	"github.com/nownabe/plummet/pkg/chibi"
	"github.com/nownabe/plummet/pkg/slack"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := initConfig()
	if err != nil {
		panic(err)
	}

	if err := chibi.InitLogger(cfg.LogLevel, cfg.LogPretty); err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}

	av := newAlphaVantage(cfg.AlphaVantageAPIKey)
	s := slack.New(cfg.SlackToken)

	ap := &app{
		symbols:      cfg.Symbols,
		slackChannel: cfg.SlackChannel,
		location:     loc,
		alphaVantage: av,
		slack:        s,
	}

	chibi.Serve(ap.handler(), cfg.Port)
}
