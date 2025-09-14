package worker

import (
	"context"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := Start(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
}
