package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/config"
)

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string                   `json:"text"`
	Username    string                   `json:"username,omitempty"`
	IconEmoji   string                   `json:"icon_emoji,omitempty"`
	Channel     string                   `json:"channel,omitempty"`
	Attachments []SlackMessageAttachment `json:"attachments,omitempty"`
}

// SlackMessageAttachment represents a Slack message attachment
type SlackMessageAttachment struct {
	Color     string `json:"color,omitempty"`
	Title     string `json:"title,omitempty"`
	Text      string `json:"text,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}

// Notifier handles sending notifications
type Notifier struct {
	config     *config.Config
	httpClient *http.Client
}

// NewNotifier creates a new notification handler
func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendSlackMessage sends a message to Slack if configured
func (n *Notifier) SendSlackMessage(message SlackMessage) error {
	if !n.config.IsSlackEnabled() {
		// Slack not configured, silently skip
		return nil
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %v", err)
	}

	resp, err := n.httpClient.Post(n.config.SlackWebhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// NotifyTrade sends a trading notification
func (n *Notifier) NotifyTrade(botID, symbol, side string, quantity, price float64) error {
	if !n.config.IsSlackEnabled() {
		return nil
	}

	message := SlackMessage{
		Text:      "🤖 Trading Signal",
		Username:  "SignalBot",
		IconEmoji: ":robot_face:",
		Attachments: []SlackMessageAttachment{
			{
				Color: func() string {
					if side == "long" {
						return "good"
					}
					return "warning"
				}(),
				Title: fmt.Sprintf("%s %s Signal", symbol, side),
				Text: fmt.Sprintf("Bot: %s\nQuantity: %.5f\nPrice: $%.2f", 
					botID, quantity, price),
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return n.SendSlackMessage(message)
}

// NotifyPrediction sends a prediction notification
func (n *Notifier) NotifyPrediction(botID, symbol, prediction string, confidence int) error {
	if !n.config.IsSlackEnabled() {
		return nil
	}

	message := SlackMessage{
		Text:      "🔮 AI Prediction",
		Username:  "SignalBot",
		IconEmoji: ":crystal_ball:",
		Attachments: []SlackMessageAttachment{
			{
				Color: func() string {
					if confidence >= 80 {
						return "good"
					} else if confidence >= 60 {
						return "warning"
					}
					return "#cccccc"
				}(),
				Title: fmt.Sprintf("%s Prediction", symbol),
				Text: fmt.Sprintf("Bot: %s\nPrediction: %s\nConfidence: %d%%", 
					botID, prediction, confidence),
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return n.SendSlackMessage(message)
}

// NotifyError sends an error notification
func (n *Notifier) NotifyError(botID, component, errorMsg string) error {
	if !n.config.IsSlackEnabled() {
		return nil
	}

	message := SlackMessage{
		Text:      "⚠️ System Error",
		Username:  "SignalBot",
		IconEmoji: ":warning:",
		Attachments: []SlackMessageAttachment{
			{
				Color:     "danger",
				Title:     fmt.Sprintf("Error in %s", component),
				Text:      fmt.Sprintf("Bot: %s\nError: %s", botID, errorMsg),
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return n.SendSlackMessage(message)
}

// IsEnabled returns true if Slack notifications are configured
func (n *Notifier) IsEnabled() bool {
	return n.config.IsSlackEnabled()
}
