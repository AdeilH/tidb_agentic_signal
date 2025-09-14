package notifications

import (
	"testing"

	"github.com/adeilh/agentic_go_signals/internal/config"
)

func TestNewNotifier(t *testing.T) {
	cfg := &config.Config{SlackWebhook: "https://hooks.slack.com/test"}
	notifier := NewNotifier(cfg)

	if notifier == nil {
		t.Fatal("expected notifier to be non-nil")
	}

	if notifier.config != cfg {
		t.Fatal("expected config to be set")
	}

	if notifier.httpClient == nil {
		t.Fatal("expected http client to be set")
	}
}

func TestIsEnabled(t *testing.T) {
	// Test with Slack enabled
	cfg := &config.Config{SlackWebhook: "https://hooks.slack.com/test"}
	notifier := NewNotifier(cfg)

	if !notifier.IsEnabled() {
		t.Fatal("expected notifier to be enabled when webhook is set")
	}

	// Test with Slack disabled
	cfg = &config.Config{SlackWebhook: ""}
	notifier = NewNotifier(cfg)

	if notifier.IsEnabled() {
		t.Fatal("expected notifier to be disabled when webhook is empty")
	}
}

func TestSendSlackMessage_Disabled(t *testing.T) {
	// Test that messages are silently skipped when Slack is disabled
	cfg := &config.Config{SlackWebhook: ""}
	notifier := NewNotifier(cfg)

	message := SlackMessage{
		Text: "Test message",
	}

	err := notifier.SendSlackMessage(message)
	if err != nil {
		t.Fatalf("expected no error when Slack is disabled, got: %v", err)
	}
}

func TestNotifyTrade_Disabled(t *testing.T) {
	cfg := &config.Config{SlackWebhook: ""}
	notifier := NewNotifier(cfg)

	err := notifier.NotifyTrade("test-bot", "BTCUSDT", "long", 0.001, 50000.0)
	if err != nil {
		t.Fatalf("expected no error when Slack is disabled, got: %v", err)
	}
}

func TestNotifyPrediction_Disabled(t *testing.T) {
	cfg := &config.Config{SlackWebhook: ""}
	notifier := NewNotifier(cfg)

	err := notifier.NotifyPrediction("test-bot", "BTCUSDT", "bullish", 85)
	if err != nil {
		t.Fatalf("expected no error when Slack is disabled, got: %v", err)
	}
}

func TestNotifyError_Disabled(t *testing.T) {
	cfg := &config.Config{SlackWebhook: ""}
	notifier := NewNotifier(cfg)

	err := notifier.NotifyError("test-bot", "predictor", "connection failed")
	if err != nil {
		t.Fatalf("expected no error when Slack is disabled, got: %v", err)
	}
}

func TestSlackMessageStructure(t *testing.T) {
	// Test that SlackMessage can be created with all fields
	message := SlackMessage{
		Text:      "Test message",
		Username:  "TestBot",
		IconEmoji: ":robot_face:",
		Channel:   "#general",
		Attachments: []SlackMessageAttachment{
			{
				Color:     "good",
				Title:     "Test Title",
				Text:      "Test attachment text",
				Timestamp: 1234567890,
			},
		},
	}

	if message.Text != "Test message" {
		t.Fatal("expected message text to be set")
	}

	if len(message.Attachments) != 1 {
		t.Fatal("expected one attachment")
	}

	if message.Attachments[0].Color != "good" {
		t.Fatal("expected attachment color to be 'good'")
	}
}
