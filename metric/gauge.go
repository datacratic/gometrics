// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"log"
	"time"
)

// Gauge defines a metric for instantaneous measurements.
type Gauge struct {
	start time.Time
	since time.Time
	sum   float64
	value float64
	valid bool
}

// Record sets the value of the gauge for supported types.
func (gauge *Gauge) Record(data interface{}) {
	now := time.Now()

	// account for the value up until now
	if gauge.valid {
		gauge.sum += gauge.value * now.Sub(gauge.since).Seconds()
	} else {
		gauge.sum = 0
		gauge.start = now
	}

	value := 0.0
	switch item := data.(type) {
	default:
		log.Panicf("unknown type %T", data)
	case bool:
		if item {
			value = 1.0
		}
	case int:
		value = float64(item)
	case int64:
		value = float64(item)
	case float64:
		value = item
	case time.Duration:
		value = item.Seconds()
	}

	gauge.value = value
	gauge.since = now
	gauge.valid = true
}

// Reset keeps the current level of the gauge but reset the partial sum.
func (gauge *Gauge) Reset() {
	now := time.Now()
	gauge.sum = 0
	gauge.start = now
	gauge.since = now
}

// Write creates a key that represents the value of the gauge over the last period of time.
func (gauge *Gauge) Write(w Writer, name string) {
	if !gauge.valid {
		return
	}

	now := time.Now()
	total := gauge.sum + gauge.value*now.Sub(gauge.since).Seconds()
	value := total / now.Sub(gauge.start).Seconds()
	w.Write(name, value)
}
