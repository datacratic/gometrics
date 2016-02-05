// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"math"
	"reflect"
	"time"
)

// Metric defines an interface for metrics of different types.
type Metric interface {
	// Record adds the specified value to the metric.
	Record(value interface{})
}

// Bool implements support for boolean metrics.
type Bool struct {
	Count int
}

// Record keeps track of the number of times the boolean value was true.
func (metric *Bool) Record(value interface{}) {
	v := value.(bool)
	if v {
		metric.Count++
	}
}

// Int implements support for numerical integral metrics.
type Int struct {
	Average float64
	Maximum int
	Minimum int
	total   int
	count   int
}

// Record keeps track of the minimum, maximum and average value of integers.
func (metric *Int) Record(value interface{}) {
	v := value.(int)
	metric.total += v

	if metric.count == 0 || v > metric.Maximum {
		metric.Maximum = v
	}

	if metric.count == 0 || v < metric.Minimum {
		metric.Minimum = v
	}

	metric.count++
	metric.Average = float64(metric.total) / float64(metric.count)
}

// Float implements support for numerical floating point metrics.
type Float struct {
	Average float64
	Maximum float64
	Minimum float64
	total   float64
	count   int
}

// Record keeps track of the minimum, maximum and average value of floating points.
func (metric *Float) Record(value interface{}) {
	v := value.(float64)
	metric.total += v

	if metric.count == 0 {
		metric.Maximum = v
		metric.Minimum = v
	}

	metric.Maximum = math.Max(v, metric.Maximum)
	metric.Minimum = math.Min(v, metric.Minimum)

	metric.count++
	metric.Average = metric.total / float64(metric.count)
}

// Duration implements support for time duration metrics.
type Duration struct {
	Average float64
	Maximum int64
	Minimum int64
	total   int64
	count   int
}

// Record keeps track of the minimum, maximum and average value of time durations.
func (metric *Duration) Record(value interface{}) {
	d := value.(time.Duration)
	v := int64(d)

	metric.total += v

	if metric.count == 0 || v > metric.Maximum {
		metric.Maximum = v
	}

	if metric.count == 0 || v < metric.Minimum {
		metric.Minimum = v
	}

	metric.count++
	metric.Average = float64(metric.total) / float64(metric.count)
}

// String implements support for string multiplicity metrics.
type String struct {
	Items map[string]int
}

// Record keeps track of the number of times a string was encountered.
func (metric *String) Record(value interface{}) {
	s := value.(string)
	if s != "" {
		v := metric.Items[s]
		v++
		metric.Items[s] = v
	}
}

// Map implements support for map of strings to metrics.
type Map struct {
	Items map[string]map[string]Metric
}

// Record keeps track of the map metrics by recursing on the value of every key.
func (metric *Map) Record(value interface{}) {
	v := reflect.ValueOf(value)
	k := v.MapKeys()
	for _, i := range k {
		name := i.Interface().(string)

		item, ok := metric.Items[name]
		if !ok {
			item = make(map[string]Metric)
			metric.Items[name] = item
		}

		recordMembers(v.MapIndex(i), item)
	}
}
