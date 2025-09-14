package svc

import (
	"sync"
	"testing"
	"github.com/hack/s/internal/config"
)

func TestConcurrentInit(t *testing.T) {
	cfg := &config.Config{KimiKey: "k", DBDSN: ":memory:"}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Init(cfg)
		}()
	}
	wg.Wait()
	if DB == nil {
		t.Fatal("expected DB initialised")
	}
}
