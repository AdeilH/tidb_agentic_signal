package worker

import (
	"context"
)

func Start(ctx context.Context, db interface{}) error {
	<-ctx.Done()
	return nil
}
