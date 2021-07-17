package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nownabe/plummet/pkg/slack"
	"github.com/rs/zerolog/log"
)

type app struct {
	symbols      []string
	slackChannel string
	slack        *slack.Client
	*alphaVantage
}

func (a *app) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)

	r.Get("/", a.handle)

	return r
}

func (a *app) handle(w http.ResponseWriter, r *http.Request) {
	attachments := []slack.Attachment{}

	for _, symbolAndMarket := range a.symbols {
		symbol := strings.Split(symbolAndMarket, ":")[0]

		ts, err := a.timeSeriesDailyAdjusted(r.Context(), symbol)
		if err != nil {
			log.Err(err).Msgf("failed to get time series of %s", symbol)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		fields := []slack.AttachmentField{}

		date := time.Now().Add(-7 * 24 * time.Hour)

		var rate float64

		for ; date.Before(time.Now()) || date.Equal(time.Now()); date = date.Add(24 * time.Hour) {
			lastWeek := date.Add(-7 * 24 * time.Hour)

			thisWeekDate := date.Format("2006-01-02")
			lastWeekDate := lastWeek.Format("2006-01-02")

			thisWeekItem, ok1 := ts.TimeSeries[thisWeekDate]
			lastWeekItem, ok2 := ts.TimeSeries[lastWeekDate]

			if !ok1 || !ok2 {
				continue
			}

			thisWeekClose, err := strconv.ParseFloat(thisWeekItem.AdjustedClose, 64)
			if err != nil {
				log.Err(err).Msg(err.Error())
				continue
			}

			lastWeekClose, err := strconv.ParseFloat(lastWeekItem.AdjustedClose, 64)
			if err != nil {
				log.Err(err).Msg(err.Error())
				continue
			}

			rate = (thisWeekClose - lastWeekClose) / lastWeekClose * 100

			fields = append([]slack.AttachmentField{{
				Title: fmt.Sprintf("%s => %s", lastWeekDate, thisWeekDate),
				Value: fmt.Sprintf("%4.2f => %4.2f (%.2f%%)", lastWeekClose, thisWeekClose, rate),
				Short: true,
			}}, fields...)
		}

		var color string
		if rate > 0 {
			color = "good"
		} else if rate <= -5 {
			color = "danger"
		} else if rate <= -3 {
			color = "warning"
		}

		attachment := slack.Attachment{
			Color:     color,
			Title:     symbol,
			TitleLink: "https://google.com/finance/quote/" + symbolAndMarket,
			Fields:    fields,
		}

		attachments = append(attachments, attachment)
	}

	req := &slack.ChatPostMessageReq{
		Channel:     a.slackChannel,
		IconEmoji:   ":chart_with_upwards_trend:",
		Username:    "Weekly Trends",
		Attachments: attachments,
		Text:        time.Now().Format("2006-01-02"),
	}

	if err := a.slack.ChatPostMessage(r.Context(), req); err != nil {
		log.Err(err).Msg("failed to post message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ok")
}
