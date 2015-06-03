// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import "time"

// Types of events recorded during the trace.
const (
	UnknownEvent = iota
	CountEvent
	SetEvent
	RecordEvent
	LogEvent
	StartEvent
	EnterEvent
	LeaveEvent
)

// Event contains the data gathered for a single event during a trace.
type Event struct {
	// From contains the index of the parent span.
	From int64
	// Kind contains the type of event.
	Kind int
	// When indicates when this event was recorded relative to the beginning of the trace.
	When time.Duration
	// What contains the name of the event.
	What string
	// Path is filled so that children can use it as a prefix to build their name.
	Path string
	// Data contains whatever relevant data is needed for this event.
	Data interface{}
}
