package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	alphaVantageHost     = "www.alphavantage.co"
	alphaVantageFunction = "TIME_SERIES_DAILY"
)

type stockMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type stockTimeSeriesItem struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
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

func (av *alphaVantage) timeSeriesDaily(
	ctx context.Context, symbol string) (*stockTimeSeries, error) {

	url := fmt.Sprintf(
		"https://%s/query?function=%s&symbol=%s&apikey=%s",
		alphaVantageHost,
		alphaVantageFunction,
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
