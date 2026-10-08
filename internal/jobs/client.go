package jobs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// Handlers bundles the callbacks invoked by job workers.
type Handlers struct {
	Notification  func(ctx context.Context, args NotificationArgs) error
	StatusCheck   func(ctx context.Context, args StatusCheckArgs) error
	LinkCheck     func(ctx context.Context, args LinkCheckArgs) error
	DataRetention func(ctx context.Context, args DataRetentionArgs) error
}

// Client wraps a River client for the pgx driver.
type Client struct {
	rc *river.Client[pgx.Tx]
}

// Migrate applies River's own schema migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return err
	}
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}

// NewClient builds a River client with the provided handlers and periodic jobs.
func NewClient(pool *pgxpool.Pool, h Handlers) (*Client, error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &notificationWorker{handler: h.Notification})
	river.AddWorker(workers, &statusCheckWorker{handler: h.StatusCheck})
	river.AddWorker(workers, &linkCheckWorker{handler: h.LinkCheck})
	river.AddWorker(workers, &dataRetentionWorker{handler: h.DataRetention})

	rc, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
		PeriodicJobs: []*river.PeriodicJob{
			river.NewPeriodicJob(
				river.PeriodicInterval(7*24*time.Hour),
				func() (river.JobArgs, *river.InsertOpts) { return LinkCheckArgs{}, nil },
				&river.PeriodicJobOpts{RunOnStart: false},
			),
			river.NewPeriodicJob(
				river.PeriodicInterval(24*time.Hour),
				func() (river.JobArgs, *river.InsertOpts) { return DataRetentionArgs{}, nil },
				&river.PeriodicJobOpts{RunOnStart: true},
			),
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{rc: rc}, nil
}

// Start begins processing jobs.
func (c *Client) Start(ctx context.Context) error { return c.rc.Start(ctx) }

// Stop gracefully stops the client.
func (c *Client) Stop(ctx context.Context) error { return c.rc.Stop(ctx) }

// InsertNotification enqueues a notification job.
func (c *Client) InsertNotification(ctx context.Context, id string) error {
	_, err := c.rc.Insert(ctx, NotificationArgs{NotificationID: id}, nil)
	return err
}

// InsertStatusCheck enqueues a status-check job.
func (c *Client) InsertStatusCheck(ctx context.Context, appID string) error {
	_, err := c.rc.Insert(ctx, StatusCheckArgs{ApplicationID: appID}, nil)
	return err
}

// InsertDataRetention enqueues a data-retention job.
func (c *Client) InsertDataRetention(ctx context.Context) error {
	_, err := c.rc.Insert(ctx, DataRetentionArgs{}, nil)
	return err
}

// InsertLinkCheck enqueues a link-check job.
func (c *Client) InsertLinkCheck(ctx context.Context) error {
	_, err := c.rc.Insert(ctx, LinkCheckArgs{}, nil)
	return err
}
