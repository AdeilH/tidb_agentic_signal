package db

import (
	"testing"
)

func TestOpen(t *testing.T) {
	// Test with invalid DSN to ensure error handling
	_, err := Open("invalid-dsn")
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

func TestAutoMigrate(t *testing.T) {
	// Test that AutoMigrate handles nil connection gracefully
	db := &DB{conn: nil}
	err := AutoMigrate(db)
	if err == nil {
		t.Fatal("expected error for nil connection")
	}
	t.Log("AutoMigrate properly failed with nil connection:", err)
}
