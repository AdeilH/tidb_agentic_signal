package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/risk"
)

// Orchestrator coordinates the full trading pipeline
type Orchestrator struct {
	db       *db.DB
	riskCalc *risk.Calculator

	// Timing configuration
	ingestInterval  time.Duration
	predictInterval time.Duration
	executeInterval time.Duration

	// State
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Bot configuration
	botID   string
	symbols []string
	enabled bool
}

// Config contains orchestrator configuration
type Config struct {
	BotID           string
	Symbols         []string
	IngestInterval  time.Duration
	PredictInterval time.Duration
	ExecuteInterval time.Duration
	RiskParams      risk.RiskParams
}

// NewOrchestrator creates a new trading orchestrator
func NewOrchestrator(cfg Config, dbConn *db.DB) (*Orchestrator, error) {
	if cfg.BotID == "" {
		return nil, fmt.Errorf("bot ID is required")
	}

	if len(cfg.Symbols) == 0 {
		cfg.Symbols = []string{"BTCUSDT", "ETHUSDT"} // Default symbols
	}

	// Initialize risk calculator
	riskCalc, err := risk.NewCalculator(cfg.RiskParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create risk calculator: %v", err)
	}

	// Set default intervals if not provided
	if cfg.IngestInterval == 0 {
		cfg.IngestInterval = 5 * time.Minute
	}
	if cfg.PredictInterval == 0 {
		cfg.PredictInterval = 10 * time.Minute
	}
	if cfg.ExecuteInterval == 0 {
		cfg.ExecuteInterval = 1 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Orchestrator{
		db:              dbConn,
		riskCalc:        riskCalc,
		ingestInterval:  cfg.IngestInterval,
		predictInterval: cfg.PredictInterval,
		executeInterval: cfg.ExecuteInterval,
		ctx:             ctx,
		cancel:          cancel,
		botID:           cfg.BotID,
		symbols:         cfg.Symbols,
		enabled:         true,
	}, nil
}

// Start begins the orchestrator pipeline
func (o *Orchestrator) Start() {
	log.Printf("Starting orchestrator for bot %s with symbols %v", o.botID, o.symbols)

	// Start ingestion worker
	o.wg.Add(1)
	go o.ingestWorker()

	// Start prediction worker
	o.wg.Add(1)
	go o.predictWorker()

	// Start execution worker
	o.wg.Add(1)
	go o.executeWorker()

	log.Println("All workers started successfully")
}

// Stop gracefully shuts down the orchestrator
func (o *Orchestrator) Stop() {
	log.Println("Stopping orchestrator...")
	o.enabled = false
	o.cancel()
	o.wg.Wait()
	log.Println("Orchestrator stopped")
}

// ingestWorker handles data ingestion at regular intervals
func (o *Orchestrator) ingestWorker() {
	defer o.wg.Done()

	ticker := time.NewTicker(o.ingestInterval)
	defer ticker.Stop()

	// Run immediately
	o.runIngestion()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			if o.enabled {
				o.runIngestion()
			}
		}
	}
}

// predictWorker generates predictions at regular intervals
func (o *Orchestrator) predictWorker() {
	defer o.wg.Done()

	ticker := time.NewTicker(o.predictInterval)
	defer ticker.Stop()

	// Wait a bit before first prediction to let ingestion run
	time.Sleep(30 * time.Second)

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			if o.enabled {
				o.runPrediction()
			}
		}
	}
}

// executeWorker handles trade execution at regular intervals
func (o *Orchestrator) executeWorker() {
	defer o.wg.Done()

	ticker := time.NewTicker(o.executeInterval)
	defer ticker.Stop()

	// Wait a bit before first execution to let predictions run
	time.Sleep(2 * time.Minute)

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			if o.enabled {
				o.runExecution()
			}
		}
	}
}

// runIngestion performs data ingestion
func (o *Orchestrator) runIngestion() {
	log.Println("Running data ingestion...")

	// Mock ingestion for now
	log.Println("Mock: Fetching news and chain metrics...")

	// Create a mock event
	event := db.Event{
		BotID:  o.botID,
		Source: "mock",
		Text:   "Mock ingestion event",
	}

	if o.db != nil && o.db.GetConn() != nil {
		// Insert using raw SQL since we don't have GORM
		query := `INSERT INTO events (bot_id, source, text, ts) VALUES (?, ?, ?, NOW())`
		_, err := o.db.GetConn().Exec(query, event.BotID, event.Source, event.Text)
		if err != nil {
			log.Printf("Failed to save mock event: %v", err)
		} else {
			log.Println("Mock ingestion event saved successfully")
		}
	}
}

// runPrediction generates new predictions
func (o *Orchestrator) runPrediction() {
	log.Println("Running prediction generation...")

	for _, symbol := range o.symbols {
		// Mock prediction generation
		prediction := db.Prediction{
			BotID:  o.botID,
			Symbol: symbol,
			Dir:    "bullish",
			Conv:   75, // 75% confidence
			Logic:  "Mock prediction: Market appears neutral",
		}

		if o.db != nil && o.db.GetConn() != nil {
			query := `INSERT INTO predictions (bot_id, symbol, dir, conv, logic, ts) VALUES (?, ?, ?, ?, ?, NOW())`
			_, err := o.db.GetConn().Exec(query, prediction.BotID, prediction.Symbol, prediction.Dir, prediction.Conv, prediction.Logic)
			if err != nil {
				log.Printf("Failed to save prediction for %s: %v", symbol, err)
				continue
			}
		}

		log.Printf("Generated mock prediction for %s (confidence: %d%%)",
			symbol, prediction.Conv)
	}
}

// runExecution evaluates predictions and executes trades
func (o *Orchestrator) runExecution() {
	log.Println("Running trade execution evaluation...")

	for _, symbol := range o.symbols {
		// Mock execution
		currentPrice := 50000.0 // Mock price
		direction := "long"     // Mock direction

		// Calculate position size
		positionSize, err := o.riskCalc.CalculatePositionSize(currentPrice, direction)
		if err != nil {
			log.Printf("Failed to calculate position size for %s: %v", symbol, err)
			continue
		}

		// Validate risk limits (assuming no current exposure for simplicity)
		err = o.riskCalc.ValidateRiskLimits(positionSize.Value, 0)
		if err != nil {
			log.Printf("Risk validation failed for %s: %v", symbol, err)
			continue
		}

		// Create mock trade record
		trade := db.Trade{
			BotID:  o.botID,
			Symbol: symbol,
			Side:   direction,
			Qty:    positionSize.Quantity,
			Price:  currentPrice,
			Status: "simulated",
		}

		// Save trade record
		if o.db != nil && o.db.GetConn() != nil {
			query := `INSERT INTO trades (bot_id, symbol, side, qty, price, status, ts) VALUES (?, ?, ?, ?, ?, ?, NOW())`
			_, err = o.db.GetConn().Exec(query, trade.BotID, trade.Symbol, trade.Side, trade.Qty, trade.Price, trade.Status)
			if err != nil {
				log.Printf("Failed to save trade record: %v", err)
			} else {
				log.Printf("Mock trade saved: %s %s %.5f @ %.2f",
					direction, symbol, positionSize.Quantity, currentPrice)
			}
		}
	}
}

// GetStatus returns the current orchestrator status
func (o *Orchestrator) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"bot_id":  o.botID,
		"symbols": o.symbols,
		"enabled": o.enabled,
		"intervals": map[string]interface{}{
			"ingest_minutes":  o.ingestInterval.Minutes(),
			"predict_minutes": o.predictInterval.Minutes(),
			"execute_minutes": o.executeInterval.Minutes(),
		},
		"risk_metrics": o.riskCalc.GetRiskMetrics(),
	}
}

// SetEnabled enables or disables the orchestrator
func (o *Orchestrator) SetEnabled(enabled bool) {
	o.enabled = enabled
	log.Printf("Orchestrator enabled: %v", enabled)
}
