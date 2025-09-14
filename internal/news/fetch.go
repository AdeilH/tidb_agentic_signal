package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Story struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Source    string `json:"source"`
	URL       string `json:"url"`
	Published string `json:"published_on"`
}

type CryptoCompareResponse struct {
	Data []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		SourceName  string `json:"source"`
		URL         string `json:"url"`
		PublishedOn int64  `json:"published_on"`
		Categories  string `json:"categories"`
	} `json:"Data"`
}

func Latest() ([]Story, error) {
	url := "https://min-api.cryptocompare.com/data/v2/news/?lang=EN&limit=50"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CryptoCompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	stories := make([]Story, len(response.Data))
	for i, item := range response.Data {
		stories[i] = Story{
			Title:     item.Title,
			Body:      item.Body,
			Source:    item.SourceName,
			URL:       item.URL,
			Published: time.Unix(item.PublishedOn, 0).Format(time.RFC3339),
		}
	}

	return stories, nil
}
