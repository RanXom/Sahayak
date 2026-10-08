package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

type notificationWorker struct {
	river.WorkerDefaults[NotificationArgs]
	handler func(ctx context.Context, args NotificationArgs) error
}

func (w *notificationWorker) Work(ctx context.Context, job *river.Job[NotificationArgs]) error {
	if w.handler == nil {
		return nil
	}
	return w.handler(ctx, job.Args)
}

type statusCheckWorker struct {
	river.WorkerDefaults[StatusCheckArgs]
	handler func(ctx context.Context, args StatusCheckArgs) error
}

func (w *statusCheckWorker) Work(ctx context.Context, job *river.Job[StatusCheckArgs]) error {
	if w.handler == nil {
		return nil
	}
	return w.handler(ctx, job.Args)
}

type linkCheckWorker struct {
	river.WorkerDefaults[LinkCheckArgs]
	handler func(ctx context.Context, args LinkCheckArgs) error
}

func (w *linkCheckWorker) Work(ctx context.Context, job *river.Job[LinkCheckArgs]) error {
	if w.handler == nil {
		return nil
	}
	return w.handler(ctx, job.Args)
}

type dataRetentionWorker struct {
	river.WorkerDefaults[DataRetentionArgs]
	handler func(ctx context.Context, args DataRetentionArgs) error
}

func (w *dataRetentionWorker) Work(ctx context.Context, job *river.Job[DataRetentionArgs]) error {
	if w.handler == nil {
		return nil
	}
	return w.handler(ctx, job.Args)
}
