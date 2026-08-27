// Package telegram is a minimal Telegram Bot API client — just enough to
// push plain notification messages into a chat (typically a group). It is
// not a general Bot API wrapper.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client posts messages to a single, pre-configured Telegram chat via a bot
// token. Create one bot with @BotFather, add it to the target group, and use
// the group's chat ID (a negative number for groups/supergroups).
type Client struct {
	baseURL string // overridable in tests; defaults to the real Bot API
	token   string
	chatID  string
	http    *http.Client
}

// NewClient builds a Client for the given bot token and destination chat ID.
func NewClient(token, chatID string) *Client {
	return &Client{
		baseURL: "https://api.telegram.org",
		token:   token,
		chatID:  chatID,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// SendMessage posts text (HTML formatting allowed, e.g. <b>/<i>) to the
// configured chat.
func (c *Client) SendMessage(ctx context.Context, text string) error {
	body, err := json.Marshal(sendMessageRequest{
		ChatID:                c.chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("telegram API error: %s", out.Description)
	}
	return nil
}
