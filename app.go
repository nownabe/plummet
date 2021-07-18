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
			httpError(w, http.StatusInternalServerError)
			return
		}

		attachments = append(attachments, attachment)
	}

	req := &slack.ChatPostMessageReq{
		Channel:     a.slackChannel,
		IconEmoji:   ":chart_with_upwards_trend:",
		Username:    "Weekly Trends",
		Attachments: attachments,
		Text:        ":chart_with_upwards_trend: *Today's fluctuation rate (" + time.Now().Format("2006-01-02") + ")* :chart_with_downwards_trend:",
	}

	if err := a.slack.ChatPostMessage(r.Context(), req); err != nil {
		log.Err(err).Msg("failed to post message")
		httpError(w, http.StatusInternalServerError)
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

	attachment := &slack.Attachment{
		Title:     symbol,
		TitleLink: "https://google.com/finance/quote/" + symbolAndMarket,
	}

	date := time.Now().In(a.location)
	dateKey := date.Format("2006-01-02")

	if _, ok := ts.TimeSeries[dateKey]; !ok {
		log.Info().Msgf("%s has no data at %s", symbolAndMarket, dateKey)
		attachment.Text = "no data"
		return attachment, nil
	}

	attachment.Text = ts.TimeSeries[dateKey].AdjustedClose
	attachment.Color = getAttachmentColor(ts, date)

	fields := []*slack.AttachmentField{}

	for i := 1; i <= 7; i++ {
		date2 := date.Add(time.Duration(-i) * 24 * time.Hour)

		field, err := a.buildAttachmentField(ts, date, date2)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to construct field for %s", symbolAndMarket)
			continue
		}

		fields = append([]*slack.AttachmentField{field}, fields...)
	}

	attachment.Fields = fields

	return attachment, nil
}

func (a *app) buildAttachmentField(
	ts *stockTimeSeries, date1t, date2t time.Time) (*slack.AttachmentField, error) {

	date1 := date1t.Format("2006-01-02")
	date2 := date2t.Format("2006-01-02")

	date1Close, date2Close, err := compare(ts, date1, date2)
	if err != nil {
		return nil, err
	}

	diffDays := int(date1t.Sub(date2t).Hours()) / 24

	rate := (date1Close - date2Close) / date2Close * 100

	var icon string

	if date1Close > date2Close {
		icon = ":arrow_upper_right:"
	} else if date1Close < date2Close {
		icon = ":arrow_lower_right:"
	} else {
		icon = ":arrow_right:"
	}

	return &slack.AttachmentField{
		Title: fmt.Sprintf("%d days (%s)", diffDays, date2),
		Value: fmt.Sprintf("%s *`%2.2f%%`* : `%4.2f` ⇨ `%4.2f`", icon, rate, date2Close, date1Close),
		Short: false,
	}, nil
}

func compare(ts *stockTimeSeries, date1, date2 string) (float64, float64, error) {
	date1Item, ok1 := ts.TimeSeries[date1]
	date2Item, ok2 := ts.TimeSeries[date2]

	if !ok1 {
		return 0, 0, fmt.Errorf("%s doesn't exist", date1)
	}

	if !ok2 {
		return 0, 0, fmt.Errorf("%s doesn't exist", date2)
	}

	date1Close, err := strconv.ParseFloat(date1Item.AdjustedClose, 64)
	if err != nil {
		log.Err(err).Msg(err.Error())
		return 0, 0, err
	}

	date2Close, err := strconv.ParseFloat(date2Item.AdjustedClose, 64)
	if err != nil {
		log.Err(err).Msg(err.Error())
		return 0, 0, err
	}

	return date1Close, date2Close, nil
}

func getAttachmentColor(ts *stockTimeSeries, date time.Time) string {
	date2 := date.Add(-7 * 24 * time.Hour)

	d1Close, d2Close, err := compare(ts, date.Format("2006-01-02"), date2.Format("2006-01-02"))
	if err != nil {
		return ""
	}

	rate := (d1Close - d2Close) / d2Close * 100

	if rate > 0 {
		return "good"
	} else if rate <= -3 {
		return "warning"
	} else if rate <= -5 {
		return "danger"
	}

	return ""
}

func httpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
