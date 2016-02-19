// Copyright (c) 2015 Datacratic. All rights reserved.

package metric

import (
	"sync"
	"sync/atomic"
)

type Aggregate struct {
	Counts map[string]int64
	Commit func(counts map[string]int64) bool

	once sync.Once
	mu   sync.RWMutex
}

func (a *Aggregate) NewWriter(s *Summary) Writer {
	a.once.Do(func() {
		a.Counts = make(map[string]int64)
	})

	return &aggregateWriter{a: a}
}

type aggregateWriter struct {
	a *Aggregate
}

func (w *aggregateWriter) Write(name string, value float64) error {
	return ErrIgnored
}

func (w *aggregateWriter) WriteScaled(name string, value float64) (err error) {
	n := int64(value)

	w.a.mu.RLock()
	count, ok := w.a.Counts[name]
	w.a.mu.RUnlock()

	if ok {
		atomic.AddInt64(&count, n)
	} else {
		w.a.mu.Lock()
		w.a.Counts[name] = n
		w.a.mu.Unlock()
	}

	return ErrIgnored
}

func (w *aggregateWriter) WriteString(name, text string) (err error) {
	return ErrIgnored
}

func (w *aggregateWriter) Close() {
	w.a.mu.Lock()

	if w.a.Commit != nil {
		done := w.a.Commit(w.a.Counts)
		if done {
			w.a.Counts = make(map[string]int64)
		}
	}

	w.a.mu.Unlock()
}
