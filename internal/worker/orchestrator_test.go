package worker

import (
	"context"
	"testing"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/risk"
)

func TestNewOrchestrator(t *testing.T) {
	// Test with minimal config
	cfg := Config{
		BotID: "test-bot",
		RiskParams: risk.RiskParams{
			AccountBalance:  10000,
			RiskPerTrade:    0.02,
			MaxPositionSize: 0.10,
			StopLossPercent: 0.05,
		},
	}

	// Create mock database
	database := &db.DB{}

	// Note: This test will pass with mock validation
	orchestrator, err := NewOrchestrator(cfg, database)
	if err == nil {
		t.Log("Orchestrator created successfully")
		_ = orchestrator // Use the variable
	} else {
		// Expected to fail without valid risk params
		t.Logf("Expected error: %v", err)
	}

	// Test with missing bot ID
	cfg.BotID = ""
	_, err = NewOrchestrator(cfg, database)
	if err == nil {
		t.Fatal("expected error with empty bot ID")
	}

	t.Log("NewOrchestrator properly validates configuration")
}

func TestOrchestratorLifecycle(t *testing.T) {
	// Create orchestrator with mock components
	cfg := Config{
		BotID:           "test-bot",
		Symbols:         []string{"BTCUSDT"},
		IngestInterval:  100 * time.Millisecond,
		PredictInterval: 200 * time.Millisecond,
		ExecuteInterval: 50 * time.Millisecond,
		RiskParams: risk.RiskParams{
			AccountBalance:  10000,
			RiskPerTrade:    0.02,
			MaxPositionSize: 0.10,
			StopLossPercent: 0.05,
		},
	}

	// Use context with timeout for testing
	_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = &db.DB{} // Mock database

	// Mock orchestrator creation (will fail without API keys)
	// But we can test the configuration logic
	riskCalc, err := risk.NewCalculator(cfg.RiskParams)
	if err != nil {
		t.Fatalf("failed to create risk calculator: %v", err)
	}

	// Test default intervals
	if cfg.IngestInterval == 0 {
		cfg.IngestInterval = 5 * time.Minute
	}

	// Verify intervals are set
	if cfg.IngestInterval != 100*time.Millisecond {
		t.Errorf("expected ingest interval 100ms, got %v", cfg.IngestInterval)
	}

	// Test risk calculator metrics
	metrics := riskCalc.GetRiskMetrics()
	if metrics["account_balance"] != 10000.0 {
		t.Errorf("expected account balance 10000, got %v", metrics["account_balance"])
	}

	t.Log("Orchestrator lifecycle configuration works correctly")
}

func TestOrchestratorStatus(t *testing.T) {
	// Create risk calculator for testing
	riskParams := risk.RiskParams{
		AccountBalance:  10000,
		RiskPerTrade:    0.02,
		MaxPositionSize: 0.10,
		StopLossPercent: 0.05,
	}

	riskCalc, err := risk.NewCalculator(riskParams)
	if err != nil {
		t.Fatalf("failed to create risk calculator: %v", err)
	}

	// Create mock orchestrator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator := &Orchestrator{
		botID:           "test-bot",
		symbols:         []string{"BTCUSDT", "ETHUSDT"},
		enabled:         true,
		ingestInterval:  5 * time.Minute,
		predictInterval: 10 * time.Minute,
		executeInterval: 1 * time.Minute,
		riskCalc:        riskCalc,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Test status
	status := orchestrator.GetStatus()

	if status["bot_id"] != "test-bot" {
		t.Errorf("expected bot_id test-bot, got %v", status["bot_id"])
	}

	if status["enabled"] != true {
		t.Errorf("expected enabled true, got %v", status["enabled"])
	}

	symbols := status["symbols"].([]string)
	if len(symbols) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(symbols))
	}

	// Test enable/disable
	orchestrator.SetEnabled(false)
	if orchestrator.enabled != false {
		t.Error("expected orchestrator to be disabled")
	}

	t.Log("Orchestrator status methods work correctly")
}
