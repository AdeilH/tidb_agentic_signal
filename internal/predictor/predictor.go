package predictor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/ingest"
	"github.com/adeilh/agentic_go_signals/internal/kimi"
)

func Generate(database *db.DB, kimiClient *kimi.Client, botID, symbol string) (*db.Prediction, error) {
	if database == nil || database.GetConn() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if kimiClient == nil {
		return nil, fmt.Errorf("kimi client is nil")
	}

	// Get recent events for context
	events, err := ingest.GetRecentEvents(database, botID, symbol, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	// Prepare context for Kimi
	newsData := buildNewsContext(events)
	chainData := buildChainContext(events)

	// Generate prediction using Kimi
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prediction, err := kimiClient.GeneratePrediction(ctx, symbol, newsData, chainData)
	if err != nil {
		// If Kimi fails, generate a conservative fallback prediction
		prediction = kimi.Prediction{
			Dir:   "FLAT",
			Conv:  30,
			Logic: "Unable to generate AI prediction, defaulting to neutral",
		}
	}

	// Store prediction in database
	now := time.Now()

	query := `
		INSERT INTO predictions (bot_id, ts, symbol, dir, conv, logic) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := database.GetConn().Exec(query, botID, now, symbol, prediction.Dir, prediction.Conv, prediction.Logic)
	if err != nil {
		return nil, fmt.Errorf("failed to insert prediction: %w", err)
	}

	id, _ := result.LastInsertId()

	dbPrediction := &db.Prediction{
		ID:     uint(id),
		BotID:  botID,
		Ts:     now,
		Symbol: symbol,
		Dir:    prediction.Dir,
		Conv:   prediction.Conv,
		Logic:  prediction.Logic,
	}

	return dbPrediction, nil
}

func buildNewsContext(events []db.Event) string {
	var newsEvents []string

	for _, event := range events {
		if event.Source == "news" && len(newsEvents) < 5 {
			// Truncate long text
			text := event.Text
			if len(text) > 200 {
				text = text[:200] + "..."
			}
			newsEvents = append(newsEvents, text)
		}
	}

	if len(newsEvents) == 0 {
		return "No recent news available"
	}

	return strings.Join(newsEvents, "\n")
}

func buildChainContext(events []db.Event) string {
	var chainEvents []string

	for _, event := range events {
		if event.Source == "chain" && len(chainEvents) < 3 {
			chainEvents = append(chainEvents, event.Text)
		}
	}

	if len(chainEvents) == 0 {
		return "No recent chain data available"
	}

	return strings.Join(chainEvents, "\n")
}

func GetLatest(database *db.DB, botID, symbol string) (*db.Prediction, error) {
	if database == nil || database.GetConn() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, bot_id, ts, symbol, dir, conv, logic, fwd_ret
		FROM predictions 
		WHERE bot_id = ? AND symbol = ? 
		ORDER BY ts DESC 
		LIMIT 1
	`

	row := database.GetConn().QueryRow(query, botID, symbol)

	var prediction db.Prediction
	err := row.Scan(&prediction.ID, &prediction.BotID, &prediction.Ts, &prediction.Symbol,
		&prediction.Dir, &prediction.Conv, &prediction.Logic, &prediction.FwdRet)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest prediction: %w", err)
	}

	return &prediction, nil
}

func GetHistory(database *db.DB, botID, symbol string, limit int) ([]db.Prediction, error) {
	if database == nil || database.GetConn() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, bot_id, ts, symbol, dir, conv, logic, fwd_ret
		FROM predictions 
		WHERE bot_id = ? AND symbol = ? 
		ORDER BY ts DESC 
		LIMIT ?
	`

	rows, err := database.GetConn().Query(query, botID, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query predictions: %w", err)
	}
	defer rows.Close()

	var predictions []db.Prediction
	for rows.Next() {
		var prediction db.Prediction
		err := rows.Scan(&prediction.ID, &prediction.BotID, &prediction.Ts, &prediction.Symbol,
			&prediction.Dir, &prediction.Conv, &prediction.Logic, &prediction.FwdRet)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prediction: %w", err)
		}
		predictions = append(predictions, prediction)
	}

	return predictions, nil
}
