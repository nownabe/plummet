package main

import (
	"context"
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
	location     *time.Location
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
	attachments := []*slack.Attachment{}

	for _, symbolAndMarket := range a.symbols {
		attachment, err := a.buildAttachment(r.Context(), symbolAndMarket)
		if err != nil {
			log.Err(err).Msgf("failed build attachment for %s", symbolAndMarket)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
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

func (a *app) buildAttachment(ctx context.Context, symbolAndMarket string) (*slack.Attachment, error) {
	symbol := strings.Split(symbolAndMarket, ":")[0]

	ts, err := a.timeSeriesDailyAdjusted(ctx, symbol)
	if err != nil {
		log.Err(err).Msgf("failed to get time series of %s", symbol)
		return nil, fmt.Errorf("failed to get time series of %s: %w", symbol, err)
	}

	fields := []*slack.AttachmentField{}

	date := time.Now().Add(-7 * 24 * time.Hour)

	for ; date.Before(time.Now()) || date.Equal(time.Now()); date = date.Add(24 * time.Hour) {
		lastWeek := date.Add(-7 * 24 * time.Hour)

		field, err := a.buildAttachmentField(ts, date, lastWeek)
		if err != nil {
			log.Warn().Err(err).Msg(err.Error())
			continue
		}

		fields = append([]*slack.AttachmentField{field}, fields...)
	}

	attachment := &slack.Attachment{
		Title:     symbol,
		TitleLink: "https://google.com/finance/quote/" + symbolAndMarket,
		Fields:    fields,
	}

	return attachment, nil
}

func (a *app) buildAttachmentField(
	timeSeries *stockTimeSeries, date1, date2 time.Time) (*slack.AttachmentField, error) {

	date1Key := date1.Format("2006-01-02")
	date2Key := date2.Format("2006-01-02")

	date1Item, ok1 := timeSeries.TimeSeries[date1Key]
	date2Item, ok2 := timeSeries.TimeSeries[date2Key]

	if !ok1 {
		return nil, fmt.Errorf("%s doesn't exist", date1Key)
	}

	if !ok2 {
		return nil, fmt.Errorf("%s doesn't exist", date2Key)
	}

	date1Close, err := strconv.ParseFloat(date1Item.AdjustedClose, 64)
	if err != nil {
		log.Err(err).Msg(err.Error())
		return nil, err
	}

	date2Close, err := strconv.ParseFloat(date2Item.AdjustedClose, 64)
	if err != nil {
		log.Err(err).Msg(err.Error())
		return nil, err
	}

	rate := (date1Close - date2Close) / date2Close * 100

	return &slack.AttachmentField{
		Title: fmt.Sprintf("%s => %s", date2Key, date1Key),
		Value: fmt.Sprintf("%4.2f => %4.2f (%.2f%%)", date2Close, date1Close, rate),
		Short: true,
	}, nil
}
