package trader

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "test-secret")
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.apiKey != "test-key" {
		t.Fatal("expected api key to be set")
	}
	if client.apiSecret != "test-secret" {
		t.Fatal("expected api secret to be set")
	}
	if client.baseURL != "https://testnet.binance.vision" {
		t.Fatal("expected correct base URL")
	}
}

func TestNewProductionClient(t *testing.T) {
	client := NewProductionClient("test-key", "test-secret")
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.apiKey != "test-key" {
		t.Fatal("expected api key to be set")
	}
	if client.apiSecret != "test-secret" {
		t.Fatal("expected api secret to be set")
	}
	if client.baseURL != "https://api.binance.com" {
		t.Fatal("expected production base URL")
	}
	if client.wsURL != "wss://stream.binance.com:9443" {
		t.Fatal("expected production WebSocket URL")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// Test testnet configuration
	testnetClient := NewClientWithConfig("test-key", "test-secret", false)
	if testnetClient.baseURL != "https://testnet.binance.vision" {
		t.Fatal("expected testnet base URL")
	}

	// Test production configuration
	prodClient := NewClientWithConfig("test-key", "test-secret", true)
	if prodClient.baseURL != "https://api.binance.com" {
		t.Fatal("expected production base URL")
	}
}

func TestSign(t *testing.T) {
	client := NewClient("test-key", "test-secret")

	query := "symbol=BTCUSDT&side=BUY&type=MARKET&quantity=0.001&timestamp=1609459200000"
	signature := client.sign(query)

	if signature == "" {
		t.Fatal("expected non-empty signature")
	}

	// Test that same input produces same signature
	signature2 := client.sign(query)
	if signature != signature2 {
		t.Fatal("expected consistent signatures")
	}
}

func TestTestConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := NewClient("test-key", "test-secret")

	// This should succeed even without valid credentials
	err := client.TestConnection()
	if err != nil {
		t.Logf("Connection test failed (expected): %v", err)
	} else {
		t.Log("Connection test succeeded")
	}
}
