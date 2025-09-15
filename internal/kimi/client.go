package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey string
	client *http.Client
}

type Prediction struct {
	Dir   string `json:"dir"`   // LONG, SHORT, FLAT
	Conv  int    `json:"conv"`  // Conviction 1-100
	Logic string `json:"logic"` // Reasoning
}

type KimiRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KimiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Ask(ctx context.Context, system, user string) (Prediction, error) {
	url := "https://api.moonshot.ai/v1/chat/completions"

	reqBody := KimiRequest{
		Model: "kimi-k2-0905-preview",
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return Prediction{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return Prediction{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Prediction{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response KimiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Prediction{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return Prediction{}, fmt.Errorf("kimi API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return Prediction{}, fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content

	// Try to parse JSON response
	var prediction Prediction
	if err := json.Unmarshal([]byte(content), &prediction); err != nil {
		// If JSON parsing fails, return a default response
		return Prediction{
			Dir:   "FLAT",
			Conv:  50,
			Logic: content,
		}, nil
	}

	return prediction, nil
}

func (c *Client) GeneratePrediction(ctx context.Context, symbol string, newsData, chainData string) (Prediction, error) {
	system := `You are an expert crypto analyst. Analyze the provided news and on-chain data for the given symbol.
Respond with a JSON object containing:
- "dir": "LONG" | "SHORT" | "FLAT" 
- "conv": conviction level 1-100
- "logic": brief reasoning (max 200 chars)

Consider market sentiment, on-chain activity, and news impact. Be conservative with high conviction levels.`

	user := fmt.Sprintf(`Symbol: %s

Recent News:
%s

On-Chain Metrics:
%s

Provide your trading signal analysis:`, symbol, newsData, chainData)

	return c.Ask(ctx, system, user)
}
