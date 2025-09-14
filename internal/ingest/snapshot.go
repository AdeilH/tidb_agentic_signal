package ingest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/chain"
	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/news"
)

func Save(database *db.DB, botID string) error {
	if database == nil || database.GetConn() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Fetch news data
	stories, err := news.Latest()
	if err != nil {
		return fmt.Errorf("failed to fetch news: %w", err)
	}

	// Fetch chain metrics
	metrics, err := chain.GetMetrics()
	if err != nil {
		return fmt.Errorf("failed to fetch chain metrics: %w", err)
	}

	now := time.Now()

	// Insert news events
	for i, story := range stories {
		if i >= 10 { // Limit to 10 stories to avoid overwhelming
			break
		}

		// Create simple vector (placeholder - in real implementation would use embeddings)
		vec := createSimpleVector(story.Title + " " + story.Body)
		vecJSON, _ := json.Marshal(vec)

		// Insert into events table
		_, err := database.GetConn().Exec(
			`INSERT INTO events (bot_id, ts, symbol, source, usd_val, text) VALUES (?, ?, ?, ?, ?, ?)`,
			botID, now, "BTC", "news", 0.0, story.Title+" "+story.Body,
		)
		if err != nil {
			return fmt.Errorf("failed to insert news event: %w", err)
		}

		// Insert into event_vecs table
		_, err = database.GetConn().Exec(
			`INSERT INTO event_vecs (id, bot_id, ts, sym, vec, text) VALUES (?, ?, ?, ?, ?, ?)`,
			i+1, botID, now, "BTC", string(vecJSON), story.Title+" "+story.Body,
		)
		if err != nil {
			return fmt.Errorf("failed to insert news vector: %w", err)
		}
	}

	// Insert chain metrics as events
	chainText := fmt.Sprintf("Active addresses: %d, Transactions: %d, Price: $%.2f",
		metrics.ActiveAddresses, metrics.TxCount, metrics.Price)

	_, err = database.GetConn().Exec(
		`INSERT INTO events (bot_id, ts, symbol, source, usd_val, text) VALUES (?, ?, ?, ?, ?, ?)`,
		botID, now, "BTC", "chain", metrics.Price, chainText,
	)
	if err != nil {
		return fmt.Errorf("failed to insert chain event: %w", err)
	}

	// Insert chain metrics vector
	vec := createSimpleVector(chainText)
	vecJSON, _ := json.Marshal(vec)

	_, err = database.GetConn().Exec(
		`INSERT INTO event_vecs (id, bot_id, ts, sym, vec, text) VALUES (?, ?, ?, ?, ?, ?)`,
		1000, botID, now, "BTC", string(vecJSON), chainText,
	)
	if err != nil {
		return fmt.Errorf("failed to insert chain vector: %w", err)
	}

	return nil
}

// createSimpleVector creates a simple vector representation of text
// In a real implementation, this would use a proper embedding model
func createSimpleVector(text string) []float64 {
	words := strings.Fields(strings.ToLower(text))
	vector := make([]float64, 128) // Simple 128-dimensional vector

	// Simple hash-based vector creation
	for i, word := range words {
		if i >= len(vector) {
			break
		}

		hash := 0
		for _, char := range word {
			hash = hash*31 + int(char)
		}

		// Normalize to [-1, 1]
		vector[i%len(vector)] += float64(hash%1000)/500.0 - 1.0
	}

	// Normalize vector
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	norm = 1.0 / (1.0 + norm) // Simple normalization

	for i := range vector {
		vector[i] *= norm
	}

	return vector
}

func GetRecentEvents(database *db.DB, botID, symbol string, limit int) ([]db.Event, error) {
	if database == nil || database.GetConn() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, bot_id, ts, symbol, source, usd_val, text 
		FROM events 
		WHERE bot_id = ? AND symbol = ? 
		ORDER BY ts DESC 
		LIMIT ?
	`

	rows, err := database.GetConn().Query(query, botID, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []db.Event
	for rows.Next() {
		var event db.Event
		err := rows.Scan(&event.ID, &event.BotID, &event.Ts, &event.Symbol, &event.Source, &event.USDVal, &event.Text)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
