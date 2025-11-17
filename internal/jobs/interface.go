package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

type BackgroundJobRunner interface {
	Enqueue(ctx context.Context, args river.JobArgs) error
}
