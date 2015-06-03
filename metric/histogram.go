// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Histogram tracks the distribution of a stream of values.
type Histogram struct {
	// Percentiles contains the list of percentiles to keep track of e.g. Percentiles["99.9th"] = 99.9.
	// When nil, the defaults will contain the 50th, 90th and 99th percentiles.
	Percentiles map[string]float64

	minimum float64
	maximum float64
	items   []float64
	sorted  bool
	count   int
	total   int
	valid   bool
}

// Record adds a sample to the stream of values for supported types.
func (histogram *Histogram) Record(data interface{}) {
	value := 0.0
	switch item := data.(type) {
	default:
		log.Panicf("unknown type %T", data)
	case int:
		value = float64(item)
	case int64:
		value = float64(item)
	case float64:
		value = item
	case time.Duration:
		value = item.Seconds()
	}

	// keep a maximum of 1K items for basic reservoir sampling
	if histogram.items == nil {
		histogram.items = make([]float64, 1000)
	}

	i, n := histogram.total, len(histogram.items)
	if i < n {
		histogram.items[i] = value
		histogram.count++
	} else {
		if i := rand.Intn(i); i < n {
			histogram.items[i] = value
		}
	}

	histogram.sorted = false

	// keep track of extremes
	if i > 0 {
		histogram.minimum = math.Min(histogram.minimum, value)
		histogram.maximum = math.Max(histogram.maximum, value)
	} else {
		histogram.minimum = value
		histogram.maximum = value
	}

	histogram.total++
	histogram.valid = true
}

// Reset invalidates all values from the histogram.
func (histogram *Histogram) Reset() {
	histogram.count = 0
	histogram.total = 0
	histogram.valid = false
}

// Write creates a key for the minimum, the maximum and percentiles.
// By default, histograms creates a key for the 50th, 90th and 99th percentiles.
func (histogram *Histogram) Write(w Writer, name string) {
	if !histogram.valid {
		return
	}

	if !histogram.sorted {
		sort.Float64s(histogram.items[:histogram.count])
		histogram.sorted = true
	}

	k := float64(histogram.count)
	percentile := func(n float64) float64 {
		return histogram.items[int(k*n)]
	}

	path := name + "."
	w.Write(path+"Minimum", histogram.minimum)
	w.Write(path+"Maximum", histogram.maximum)

	if histogram.Percentiles == nil {
		w.Write(path+"50th", percentile(0.5))
		w.Write(path+"90th", percentile(0.9))
		w.Write(path+"99th", percentile(0.99))
	} else {
		for id, value := range histogram.Percentiles {
			if value < 0.0 || value > 1.0 {
				log.Fatalf("metric: invalid percentile '%s=%f' for histogram '%s'\n", id, value, name)
			}

			w.Write(path+id, percentile(value/100.0))
		}
	}
}
