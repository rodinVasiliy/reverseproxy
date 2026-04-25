package action

import "time"

type Effect struct {
	Block bool

	Logs []LogEntry

	AddToBL *BLRequest
}

type LogEntry struct {
	Message string
	Fields  map[string]string
}

type BLRequest struct {
	IP  string
	TTL time.Duration
}
