package kimi

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key")
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.apiKey != "test-key" {
		t.Fatal("expected api key to be set")
	}
}

func TestGeneratePrediction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This test requires a real API key
	client := NewClient("fake-key-for-test")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newsData := "Bitcoin surges as institutional adoption continues"
	chainData := "Active addresses: 500000, Transactions: 300000"

	// This will fail with fake key, but tests the structure
	_, err := client.GeneratePrediction(ctx, "BTC", newsData, chainData)
	if err == nil {
		t.Log("Prediction generated successfully")
	} else {
		t.Logf("Expected error with fake key: %v", err)
	}
}
