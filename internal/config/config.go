package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KimiKey       string
	BinanceKey    string
	BinanceSecret string
	DBDSN         string
	SlackWebhook  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	c := &Config{
		KimiKey:       os.Getenv("KIMI_API_KEY"),
		BinanceKey:    os.Getenv("BINANCE_TEST_KEY"),
		BinanceSecret: os.Getenv("BINANCE_TEST_SECRET"),
		DBDSN:         os.Getenv("TIDB_DSN"),
		SlackWebhook:  os.Getenv("SLACK_WEBHOOK_URL"),
	}

	// Set defaults
	if c.DBDSN == "" {
		c.DBDSN = "root:@tcp(localhost:4000)/sigforge?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Validate required fields
	if c.KimiKey == "" {
		return nil, errors.New("KIMI_API_KEY is required")
	}
	if c.BinanceKey == "" {
		return nil, errors.New("BINANCE_TEST_KEY is required")
	}
	if c.BinanceSecret == "" {
		return nil, errors.New("BINANCE_TEST_SECRET is required")
	}

	return c, nil
}

// IsSlackEnabled returns true if Slack webhook URL is configured
func (c *Config) IsSlackEnabled() bool {
	return c.SlackWebhook != ""
}
