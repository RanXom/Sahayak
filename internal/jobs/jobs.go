// Package jobs defines River job types and workers.
package jobs

// NotificationArgs dispatches a queued notification.
type NotificationArgs struct {
	NotificationID string `json:"notification_id"`
}

// Kind implements river.JobArgs.
func (NotificationArgs) Kind() string { return "notification" }

// StatusCheckArgs checks the external status of an application.
type StatusCheckArgs struct {
	ApplicationID string `json:"application_id"`
}

// Kind implements river.JobArgs.
func (StatusCheckArgs) Kind() string { return "status_check" }

// LinkCheckArgs sweeps published scheme application URLs for broken links.
type LinkCheckArgs struct{}

// Kind implements river.JobArgs.
func (LinkCheckArgs) Kind() string { return "link_check" }

// DataRetentionArgs cleans up expired data (OTPs, temp files).
type DataRetentionArgs struct{}

// Kind implements river.JobArgs.
func (DataRetentionArgs) Kind() string { return "data_retention" }
