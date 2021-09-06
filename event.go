package channelx

import (
	"time"
)

//EventID is type of events
type EventID int64

// Name returns the string name for the passed eventID
func (id EventID) Name() string {
	if eventName, ok := eventNames[id]; ok {
		return eventName
	}

	return ""
}

var eventNames = map[EventID]string{
}

// RegisterEventName registers event name
func RegisterEventName(eventID EventID, name string)  {
	eventNames[eventID] = name
}

//Event ...
type Event interface {
	ID() EventID
}

//EventHandler ...
type EventHandler interface {
	OnEvent(event Event) error
	Logger() Logger
	CanAutoRetry(err error) bool
}

// JobStatus holds information related to a job status
type JobStatus struct {
	RunAt      time.Time
	FinishedAt time.Time
	Err        error
}
