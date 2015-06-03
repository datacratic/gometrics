// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"log"
	"time"
)

// Counter defines a monitonically increasing value reported per second.
type Counter struct {
	value float64
	valid bool
}

// Record increases the value of the counter for supported types.
func (counter *Counter) Record(data interface{}) {
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

	if value < 0 {
		log.Printf("counter increment of '%f' is not monotonically-increasing\n", value)
		return
	}

	counter.value += value
	counter.valid = true
}

// Reset invalidates and restarts the counter from zero.
func (counter *Counter) Reset() {
	counter.value = 0
	counter.valid = false
}

// Write creates a key that represents the value of the counter scaled over time.
func (counter *Counter) Write(w Writer, name string) {
	if !counter.valid {
		return
	}

	w.WriteScaled(name, counter.value)
}
