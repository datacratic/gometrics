// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
)

type timeline struct {
	first span
	begin time.Time
	epoch int64
	count int64
	total int64
	mu    sync.Mutex
	queue []Event

	tracing string
	Handler
}

type span struct {
	id    int64
	epoch int64
	done  int64
	owner *timeline
	up    *span
}

// global counter for epoch
var epoch int64

// keep a pool of timelines to reduce GC pressure
var timelines = sync.Pool{
	New: func() interface{} {
		t := new(timeline)
		t.total = 4096
		t.queue = make([]Event, t.total)
		t.first.owner = t
		return t
	},
}

// spanKey is a private type to find the inner span
type spanKey int

// Tracing returns the key currently being traced as specified on Start.
func Tracing(c context.Context) string {
	s, ok := c.Value(spanKey(0)).(*span)
	if !ok {
		return ""
	}

	return s.owner.tracing
}

// Start creates a new span.
func Start(c context.Context, name, tracing string) context.Context {
	return create(c, name, tracing, StartEvent)
}

// Enter creates a new span that includes it's parent's name as part of its identifier.
func Enter(c context.Context, name string) context.Context {
	return create(c, name, "*", EnterEvent)
}

func create(c context.Context, name, tracing string, kind int) context.Context {
	s, ok := c.Value(spanKey(0)).(*span)
	if !ok {
		// is there an handler?
		handler, ok := c.Value(handlerKey(0)).(Handler)
		if !ok {
			handler = &nilHandler{}
		}

		// get the timeline storage from the pool
		t := timelines.Get().(*timeline)
		t.begin = time.Now()
		t.count = 0
		t.Handler = handler
		t.tracing = tracing

		// reset the top span
		s = &t.first
		s.done = 0
		s.epoch = t.epoch
	}

	t := s.owner

	// lock the timeline to add an event
	t.mu.Lock()

	// ignore?
	if s.epoch != t.epoch {
		t.mu.Unlock()
		return c
	}

	// record
	id := t.add(s.id, kind, name, nil)

	// we're done, unlock the timeline
	t.mu.Unlock()

	// track the span in a new context
	return context.WithValue(c, spanKey(0), &span{
		id:    id,
		epoch: s.epoch,
		owner: t,
		up:    s,
	})
}

// Leave ends the context's current span.
// A span can only be left once.
// Leaving the topmost span will flush the events of the timeline.
func Leave(c context.Context, name string) {
	s, ok := c.Value(spanKey(0)).(*span)
	if !ok {
		log.Panic("no span", name)
	}

	t := s.owner

	// lock the timeline to add an event
	t.mu.Lock()

	// ignore?
	if s.epoch != t.epoch {
		t.mu.Unlock()
		return
	}

	if s.done != 0 {
		log.Panic("span was already left")
	} else {
		s.done++
	}

	// record
	t.add(s.id, LeaveEvent, name, nil)

	// leaving the top span?
	if s.up.id == 0 {
		t.epoch = atomic.AddInt64(&epoch, 1)

		// return the timeline to the pool when done
		h := chainHandler{
			Handler: t.Handler,
			f: func() {
				timelines.Put(t)
			},
		}

		// handle events
		h.HandleTrace(t.queue[:t.count+1])
	}

	// we're done, unlock the timeline
	t.mu.Unlock()
}

// Count increases a counter.
func Count(c context.Context, name string, data interface{}) {
	store(c, name, data, CountEvent)
}

// Set stores an instantaneous measurement.
func Set(c context.Context, name string, data interface{}) {
	store(c, name, data, SetEvent)
}

// Record stores the sample value of a stream.
func Record(c context.Context, name string, data interface{}) {
	store(c, name, data, RecordEvent)
}

// Log stores values directly.
func Log(c context.Context, name string, data interface{}) {
	store(c, name, data, LogEvent)
}

func store(c context.Context, name string, data interface{}, kind int) {
	s, ok := c.Value(spanKey(0)).(*span)
	if !ok {
		log.Panic("no span")
	}

	t := s.owner

	// lock the timeline to add an event
	t.mu.Lock()

	// ignore?
	if s.epoch != t.epoch {
		t.mu.Unlock()
		return
	}

	// record
	t.add(s.id, kind, name, data)

	// we're done, unlock the timeline
	t.mu.Unlock()
}

func (t *timeline) add(from int64, kind int, what string, data interface{}) int64 {
	t.count++
	i, n := t.count, t.total

	// double the size of the queue
	if i == n {
		n = n * 2
		q := make([]Event, n)
		copy(q, t.queue)
		t.queue = q
		t.total = n
	}

	item := &t.queue[i]
	item.From = from
	item.Kind = kind
	item.When = time.Since(t.begin)
	item.What = what
	item.Path = ""
	item.Data = data

	return i
}
