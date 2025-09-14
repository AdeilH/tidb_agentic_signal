package api

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adeilh/agentic_go_signals/internal/db"
)

func TestNew(t *testing.T) {
	app := New(&db.DB{})
	if app == nil {
		t.Fatal("expected app to be non-nil")
	}
}

func TestHealthCheck(t *testing.T) {
	app := New(&db.DB{})

	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := app.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Fatal("expected 'ok' in response")
	}
}

func TestCreateBot(t *testing.T) {
	app := New(&db.DB{})

	req := httptest.NewRequest("POST", "/bot/create", nil)
	resp, err := app.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "bot_") {
		t.Fatal("expected bot ID in response")
	}
}

func TestManualIngest(t *testing.T) {
	app := New(&db.DB{})

	req := httptest.NewRequest("POST", "/ingest/manual?bot_id=test123", nil)
	resp, err := app.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
