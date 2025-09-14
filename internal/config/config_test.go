package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required env vars for test
	os.Setenv("KIMI_API_KEY", "test-kimi-key")
	os.Setenv("BINANCE_TEST_KEY", "test-binance-key")
	os.Setenv("BINANCE_TEST_SECRET", "test-binance-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if cfg.KimiKey != "test-kimi-key" {
		t.Fatal("expected KIMI_API_KEY to be set")
	}
	if cfg.BinanceKey != "test-binance-key" {
		t.Fatal("expected BINANCE_TEST_KEY to be set")
	}
	if cfg.DBDSN == "" {
		t.Fatal("expected DBDSN to have default value")
	}
}

func TestLoadMissingRequired(t *testing.T) {
	// Clear env vars
	os.Unsetenv("KIMI_API_KEY")
	os.Unsetenv("BINANCE_TEST_KEY")
	os.Unsetenv("BINANCE_TEST_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required env vars")
	}
}

func TestIsSlackEnabled(t *testing.T) {
	// Test with Slack webhook URL set
	config := &Config{SlackWebhook: "https://hooks.slack.com/test"}
	if !config.IsSlackEnabled() {
		t.Fatal("expected Slack to be enabled when webhook URL is set")
	}
	
	// Test with empty Slack webhook URL
	config = &Config{SlackWebhook: ""}
	if config.IsSlackEnabled() {
		t.Fatal("expected Slack to be disabled when webhook URL is empty")
	}
}
