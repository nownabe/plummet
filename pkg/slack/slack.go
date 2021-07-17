package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const apiURLBase = "https://slack.com/api/"

type Client struct {
	token  string
	client *http.Client
}

func New(token string) *Client {
	return &Client{
		token:  token,
		client: &http.Client{},
	}
}

type Attachment struct {
	Color     string            `json:"color,omitempty"`
	Fallback  string            `json:"fallback,omitempty"`
	Fields    []AttachmentField `json:"fields,omitempty"`
	Text      string            `json:"text,omitempty"`
	Title     string            `json:"title,omitempty"`
	TitleLink string            `json:"title_link,omitempty"`
}

type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type apiResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type ChatPostMessageReq struct {
	Channel     string       `json:"channel,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func (c *Client) ChatPostMessage(ctx context.Context, r *ChatPostMessageReq) error {
	reqBody, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL("chat.postMessage"), bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("chat.postMessage failed (%d): %s", resp.StatusCode, body)
	}

	var res apiResponse

	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	if !res.OK {
		return fmt.Errorf("chat.postMessage error: %s", res.Error)
	}

	return nil
}

func apiURL(method string) string {
	return apiURLBase + method
}
