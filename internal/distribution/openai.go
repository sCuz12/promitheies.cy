package distribution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
		http:    &http.Client{Timeout: 45 * time.Second},
	}
}

type responseRequest struct {
	Model           string `json:"model"`
	Instructions    string `json:"instructions"`
	Input           string `json:"input"`
	MaxOutputTokens int    `json:"max_output_tokens"`
	Store           bool   `json:"store"`
}

type responseBody struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) GeneratePost(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is not set")
	}
	if c.model == "" {
		return "", errors.New("OpenAI model is not set")
	}

	body, err := json.Marshal(responseRequest{
		Model:           c.model,
		Instructions:    "You write factual, awareness-focused social posts for Promitheies.cy. Return only the requested numbered post options.",
		Input:           prompt,
		MaxOutputTokens: 320,
		Store:           false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build OpenAI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send OpenAI request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read OpenAI response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("OpenAI API returned %s: %s", resp.Status, strings.TrimSpace(string(respBytes)))
	}

	var out responseBody
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("decode OpenAI response: %w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", errors.New(out.Error.Message)
	}
	if out.OutputText != "" {
		return out.OutputText, nil
	}
	for _, item := range out.Output {
		for _, content := range item.Content {
			if content.Text != "" {
				return content.Text, nil
			}
		}
	}
	return "", errors.New("OpenAI response did not include text")
}
