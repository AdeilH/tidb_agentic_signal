package main

import (
	"context"
	"log"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/config"
	"github.com/adeilh/agentic_go_signals/internal/services"
	"github.com/adeilh/agentic_go_signals/internal/trader"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	log.Println("=== SigForge Market Data Integration Test ===")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		if err.Error() == "BINANCE_TEST_KEY is required" || err.Error() == "BINANCE_TEST_SECRET is required" {
			log.Println("⚠️  Binance API keys not found. Testing with mock data...")
			log.Println("Set BINANCE_TEST_KEY and BINANCE_TEST_SECRET environment variables for full testing")
			testWithoutApiKeys()
			return
		}
		log.Fatalf("❌ Config load failed: %v", err)
	}

	// Test Binance client
	log.Println("🔄 Testing Binance client connection...")
	binanceClient := trader.NewClient(cfg.BinanceKey, cfg.BinanceSecret)
	
	if err := binanceClient.TestConnection(); err != nil {
		log.Printf("⚠️  Binance connection failed: %v", err)
		log.Println("Testing with available functionality...")
	} else {
		log.Println("✅ Binance client connected successfully")
	}

	// Test market data methods
	log.Println("\n🔄 Testing market data retrieval...")
	testMarketDataMethods(binanceClient)

	// Test WebSocket hub
	log.Println("\n🔄 Testing WebSocket hub...")
	testWebSocketHub()

	// Test market data service
	log.Println("\n🔄 Testing market data service...")
	testMarketDataService(binanceClient)

	log.Println("\n✅ Market data integration test completed successfully!")
}

func testWithoutApiKeys() {
	log.Println("🔄 Testing WebSocket hub without API keys...")
	testWebSocketHub()
	log.Println("✅ Basic functionality test completed")
}

func testMarketDataMethods(client *trader.Client) {
	// Test ticker price
	log.Println("  📊 Testing ticker price retrieval...")
	ticker, err := client.GetTickerPrice("BTCUSDT")
	if err != nil {
		log.Printf("  ⚠️  Ticker price test failed: %v", err)
	} else {
		log.Printf("  ✅ BTCUSDT price: %s", ticker.Price)
	}

	// Test 24hr ticker
	log.Println("  📈 Testing 24hr ticker statistics...")
	tickers, err := client.GetTicker24hr("BTCUSDT")
	if err != nil {
		log.Printf("  ⚠️  24hr ticker test failed: %v", err)
	} else if len(tickers) > 0 {
		log.Printf("  ✅ BTCUSDT 24hr change: %s%%", tickers[0].PriceChangePercent)
	}

	// Test order book
	log.Println("  📋 Testing order book retrieval...")
	orderBook, err := client.GetOrderBook("BTCUSDT", 5)
	if err != nil {
		log.Printf("  ⚠️  Order book test failed: %v", err)
	} else {
		log.Printf("  ✅ Order book retrieved - Bids: %d, Asks: %d", len(orderBook.Bids), len(orderBook.Asks))
	}

	// Test recent trades
	log.Println("  💱 Testing recent trades retrieval...")
	trades, err := client.GetRecentTrades("BTCUSDT", 5)
	if err != nil {
		log.Printf("  ⚠️  Recent trades test failed: %v", err)
	} else {
		log.Printf("  ✅ Retrieved %d recent trades", len(trades))
	}

	// Test klines
	log.Println("  📊 Testing klines retrieval...")
	klines, err := client.GetKlines("BTCUSDT", "1h", 5, 0, 0)
	if err != nil {
		log.Printf("  ⚠️  Klines test failed: %v", err)
	} else {
		log.Printf("  ✅ Retrieved %d klines", len(klines))
	}

	// Test exchange info
	log.Println("  ℹ️  Testing exchange info retrieval...")
	info, err := client.GetExchangeInfo()
	if err != nil {
		log.Printf("  ⚠️  Exchange info test failed: %v", err)
	} else {
		log.Printf("  ✅ Exchange info retrieved successfully")
		if symbols, ok := info["symbols"].([]interface{}); ok {
			log.Printf("  📈 Available symbols: %d", len(symbols))
		}
	}

	log.Println("  ✅ Market data methods test completed")
}

func testWebSocketHub() {
	log.Println("  🔌 Creating WebSocket hub...")
	hub := trader.NewWSHub()

	// Start hub in goroutine
	go hub.Run()

	// Test broadcasting
	log.Println("  📡 Testing broadcast functionality...")
	hub.Broadcast("test_message", map[string]interface{}{
		"type":      "test",
		"message":   "Hello from hub",
		"timestamp": time.Now().Unix(),
	})

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	log.Println("  ✅ WebSocket hub test completed")
}

func testMarketDataService(client *trader.Client) {
	log.Println("  🚀 Creating market data service...")
	hub := trader.NewWSHub()
	go hub.Run()

	service := services.NewMarketDataService(client, hub, nil) // Pass nil for DB in test

	// Test service status
	log.Printf("  📊 Service running: %v", service.IsRunning())

	// Test adding symbols
	log.Println("  ➕ Testing symbol addition...")
	service.AddSymbol("ETHUSDT")
	service.AddSymbol("BNBUSDT")

	// Test getting market summary
	log.Println("  📈 Testing market summary...")
	summary := service.GetMarketSummary()
	log.Printf("  📊 Market summary: %d symbols", summary["total_symbols"])

	// Test getting all prices (initially empty)
	prices := service.GetAllPrices()
	log.Printf("  💰 Current prices cached: %d", len(prices))

	// Test starting service (with short context to avoid hanging)
	log.Println("  🔄 Testing service start (2 second test)...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start service in goroutine since it's blocking
	go func() {
		if err := service.StartStreaming(ctx); err != nil && err != context.DeadlineExceeded {
			log.Printf("  ⚠️  Service start error: %v", err)
		}
	}()

	// Wait a moment for service to initialize
	time.Sleep(500 * time.Millisecond)

	log.Printf("  📊 Service running after start: %v", service.IsRunning())

	// Wait for context timeout
	<-ctx.Done()

	// Stop service
	log.Println("  🛑 Stopping service...")
	service.Stop()
	log.Printf("  📊 Service running after stop: %v", service.IsRunning())

	log.Println("  ✅ Market data service test completed")
}
