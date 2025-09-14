package chain

import (
	"testing"
)

func TestActiveAddr(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	count, err := ActiveAddr()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if count <= 0 {
		t.Errorf("expected positive active address count, got: %d", count)
	}

	t.Logf("Active addresses: %d", count)
}

func TestTxCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	count, err := TxCount()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if count <= 0 {
		t.Errorf("expected positive transaction count, got: %d", count)
	}

	t.Logf("Transaction count: %d", count)
}

func TestGetMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	metrics, err := GetMetrics()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if metrics.ActiveAddresses <= 0 {
		t.Error("expected positive active addresses")
	}

	if metrics.TxCount <= 0 {
		t.Error("expected positive transaction count")
	}

	if metrics.Timestamp <= 0 {
		t.Error("expected positive timestamp")
	}

	t.Logf("Metrics: ActiveAddr=%d, TxCount=%d, Price=%f, Timestamp=%d",
		metrics.ActiveAddresses, metrics.TxCount, metrics.Price, metrics.Timestamp)
}
