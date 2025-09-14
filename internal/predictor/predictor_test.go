package predictor

import (
	"testing"

	"github.com/adeilh/agentic_go_signals/internal/db"
)

func TestGenerate(t *testing.T) {
	// Test with nil database
	_, err := Generate(nil, nil, "test-bot", "BTC")
	if err == nil {
		t.Fatal("expected error with nil database")
	}

	// Test with nil kimi client
	database := &db.DB{}
	_, err = Generate(database, nil, "test-bot", "BTC")
	if err == nil {
		t.Fatal("expected error with nil kimi client")
	}

	t.Log("Generate function properly validates inputs")
}

func TestBuildNewsContext(t *testing.T) {
	events := []db.Event{
		{Source: "news", Text: "Bitcoin hits new high"},
		{Source: "news", Text: "Institutional adoption growing"},
		{Source: "chain", Text: "Active addresses: 500000"},
	}

	context := buildNewsContext(events)

	if context == "No recent news available" {
		t.Fatal("expected news context to be built")
	}

	if !contains(context, "Bitcoin hits new high") {
		t.Fatal("expected news content in context")
	}

	t.Logf("News context: %s", context)
}

func TestBuildChainContext(t *testing.T) {
	events := []db.Event{
		{Source: "news", Text: "Bitcoin hits new high"},
		{Source: "chain", Text: "Active addresses: 500000"},
		{Source: "chain", Text: "Transactions: 300000"},
	}

	context := buildChainContext(events)

	if context == "No recent chain data available" {
		t.Fatal("expected chain context to be built")
	}

	if !contains(context, "Active addresses") {
		t.Fatal("expected chain content in context")
	}

	t.Logf("Chain context: %s", context)
}

func TestGetLatest(t *testing.T) {
	// Test with nil database
	_, err := GetLatest(nil, "test-bot", "BTC")
	if err == nil {
		t.Fatal("expected error with nil database")
	}

	t.Log("GetLatest properly validates database connection")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
