// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"strings"
	"time"
)

// Metric implements an aggregation of values over a period of time.
type Metric interface {
	// Record updates the metric with the specified value.
	// Counter, Gauge and Histograms record numbers while Labels records any type as strings.
	// Supported numerical types are 'int', 'int64', 'float64', 'bool' and 'time.Duration'.
	// Boolean values are converted to 0.0 and 1.0 while durations are converted to seconds.
	Record(value interface{})
	// Reset prepares the metric for the next period of aggregation.
	Reset()
	// Write reports the aggregated metrics.
	Write(w Writer, name string)
}

// Summary represents multiple aggregations of metrics over a period of time.
type Summary struct {
	// Name contains a friendly identifier for this set of metrics.
	Name string
	// Data contains named metrics.
	Keys map[string]Metric
	// Time contains the publication time stamp.
	Time time.Time
	// Step contains the duration of the aggreation period.
	Step time.Duration
}

// Write goes over each aggregated metric and writes its value to the reporter's writer.
func (summary *Summary) Write(r Reporter) {
	w := r.NewWriter(summary)

	path := summary.Name
	if path != "" && !strings.HasSuffix(path, ".") {
		path += "."
	}

	for name, item := range summary.Keys {
		item.Write(w, path+name)
	}

	w.Close()
}

// Reset goes over each aggregated metric and reset its state.
func (summary *Summary) Reset() {
	for _, item := range summary.Keys {
		item.Reset()
	}
}

// Count updates a Counter metric.
func (summary *Summary) Count(name string, value interface{}) {
	item, ok := summary.Keys[name]
	if !ok {
		item = summary.create(name, new(Counter))
	}

	item.Record(value)
}

// Set updates a Gauge metric.
func (summary *Summary) Set(name string, value interface{}) {
	item, ok := summary.Keys[name]
	if !ok {
		item = summary.create(name, new(Gauge))
	}

	item.Record(value)
}

// Record updates an Histogram metric.
func (summary *Summary) Record(name string, value interface{}) {
	item, ok := summary.Keys[name]
	if !ok {
		item = summary.create(name, new(Histogram))
	}

	item.Record(value)
}

// Log updates a Labels metric.
func (summary *Summary) Log(name string, value interface{}) {
	item, ok := summary.Keys[name]
	if !ok {
		item = summary.create(name, new(Labels))
	}

	item.Record(value)
}

func (summary *Summary) create(name string, item Metric) Metric {
	if summary.Keys == nil {
		summary.Keys = make(map[string]Metric)
	}

	summary.Keys[name] = item
	return item
}
