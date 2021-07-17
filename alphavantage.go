package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	alphaVantageHost = "www.alphavantage.co"
)

type stockMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type stockTimeSeriesItem struct {
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

type stockTimeSeries struct {
	Metadata   *stockMetadata                  `json:"Meta Data"`
	TimeSeries map[string]*stockTimeSeriesItem `json:"Time Series (Daily)"`
}

type alphaVantage struct {
	apiKey     string
	httpClient *http.Client
}

func newAlphaVantage(apiKey string) *alphaVantage {
	return &alphaVantage{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (av *alphaVantage) timeSeriesDailyAdjusted(
	ctx context.Context, symbol string) (*stockTimeSeries, error) {

	url := fmt.Sprintf(
		"https://%s/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&apikey=%s",
		alphaVantageHost,
		symbol,
		av.apiKey,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed http.NewRequest: %w", err)
	}

	resp, err := av.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed av.httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed io.ReadAll: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("alphaVantage API error (%d): %s", resp.StatusCode, body)
	}

	ts := &stockTimeSeries{}

	if err := json.Unmarshal(body, &ts); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return ts, nil
}
