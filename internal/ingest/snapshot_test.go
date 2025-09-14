package ingest

import (
	"testing"

	"github.com/adeilh/agentic_go_signals/internal/db"
)

func TestSave(t *testing.T) {
	// Test with nil database
	err := Save(nil, "test-bot")
	if err == nil {
		t.Fatal("expected error with nil database")
	}

	// Test with empty database
	database := &db.DB{}
	err = Save(database, "test-bot")
	if err == nil {
		t.Fatal("expected error with nil connection")
	}

	t.Log("Save function properly validates database connection")
}

func TestCreateSimpleVector(t *testing.T) {
	text := "bitcoin price rising bullish market"
	vector := createSimpleVector(text)

	if len(vector) != 128 {
		t.Fatalf("expected vector length 128, got %d", len(vector))
	}

	// Check that vector is not all zeros
	hasNonZero := false
	for _, v := range vector {
		if v != 0 {
			hasNonZero = true
			break
		}
	}

	if !hasNonZero {
		t.Fatal("expected vector to have non-zero values")
	}

	t.Logf("Generated vector with %d dimensions", len(vector))
}

func TestGetRecentEvents(t *testing.T) {
	// Test with nil database
	_, err := GetRecentEvents(nil, "test-bot", "BTC", 10)
	if err == nil {
		t.Fatal("expected error with nil database")
	}

	// Test with empty database
	database := &db.DB{}
	_, err = GetRecentEvents(database, "test-bot", "BTC", 10)
	if err == nil {
		t.Fatal("expected error with nil connection")
	}

	t.Log("GetRecentEvents properly validates database connection")
}
