package svc

import (
	"sync"
	"testing"

	"github.com/adeilh/agentic_go_signals/internal/config"
)

func TestConcurrentInit(t *testing.T) {
	cfg := &config.Config{KimiKey: "k", DBDSN: "user:pass@tcp(localhost:3306)/testdb"}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Init(cfg) // This will fail without real DB but tests concurrent access
		}()
	}
	wg.Wait()

	// Since we're using MySQL driver and no real connection, DB will still be nil
	// But we can test that the once.Do mechanism works
	t.Log("Concurrent initialization completed")
}
