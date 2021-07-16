package main

import (
	"github.com/nownabe/plummet/pkg/chibi"
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

	app, err := newApp()
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}

	chibi.Serve(app.handler(), cfg.Port)
}
