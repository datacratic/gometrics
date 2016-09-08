// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"sync"
	"time"
)

// Periodic serializes access to the trace handler with a periodic report.
type Periodic struct {
	Handler
	Period time.Duration

	feed chan func()
	once sync.Once
}

// HandleTrace pushes the processing of the trace of events on the same channel as the periodic callback.
func (h *Periodic) HandleTrace(events []Event) {
	h.once.Do(h.initialize)

	done := make(chan struct{})
	h.feed <- func() {
		h.Handler.HandleTrace(events)
		close(done)
	}

	<-done
}

func (h *Periodic) Report(dt time.Duration) {
	h.once.Do(h.initialize)

	h.feed <- func() {
		h.Handler.Report(dt)
	}
}

func (h *Periodic) Close() {
	h.once.Do(h.initialize)

	h.feed <- func() {
		h.Handler.Close()
	}
}

func (h *Periodic) initialize() {
	h.feed = make(chan func(), 65536)
	go func() {
		for f := range h.feed {
			f()
		}
	}()

	dt := h.Period
	if 0 == dt {
		dt = time.Second
	}

	ticker := time.Tick(dt)
	go func() {
		for _ = range ticker {
			h.Report(dt)
		}
	}()
}
