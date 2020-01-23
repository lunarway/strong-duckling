package whooping

import "time"

type Whoop struct {
	Message      string
	From         string
	Timestamp    time.Time
	RemoteStatus WhoopRemoteStatus
}

type WhoopRemoteStatus struct {
	Open    bool
	Latency time.Duration
	Drift   time.Duration
}
