// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"reflect"
	"time"
)

// Metrics represents the aggregation of a set of metrics over a period of time.
type Metrics struct {
	// Keys contains the current set of metrics.
	Keys map[string]Metric
	// Hits contains the number of accumulated samples.
	Hits int
}

// Summary represents multiple aggregations of metrics over a period of time.
type Summary struct {
	// Name contains a friendly identifier for this set of metrics.
	Name string
	// Data contains a collection of named metrics.
	Data map[string]*Metrics
	// Send contains a sequential number of that summary based on the number of times it was published by the monitor.
	Send int
	// Time contains the time stamp of the summary publication.
	Time time.Time
	// Step contains the duration of the aggreation period.
	Step time.Duration
}

// Record uses reflection to create and aggregate metrics based on their data type.
func (summary *Summary) Record(name string, data interface{}) {
	item, ok := summary.Data[name]
	if !ok {
		item = new(Metrics)
		item.Keys = make(map[string]Metric)
		summary.Data[name] = item
	}

	item.Hits++
	recordMembers(reflect.ValueOf(data).Elem(), item.Keys, "")
}

func recordMembers(value reflect.Value, keys map[string]Metric, prefix string) {
	t := value.Type()

	if t.Kind() == reflect.Ptr {
		recordMembers(value.Elem(), keys, prefix)
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		v := value.Field(i)
		k := v.Kind()

		m, ok := keys[prefix+f.Name]
		if !ok {
			switch v.Interface().(type) {
			case bool:
				m = new(Bool)
			case int:
				m = new(Int)
			case float64:
				m = new(Float)
			case time.Duration:
				m = new(Duration)
			case string:
				m = &String{
					Items: make(map[string]int),
				}
			default:
				switch {
				case k == reflect.Map:
					m = &Map{
						Items: make(map[string]map[string]Metric),
					}
				case k == reflect.Struct:
					recordMembers(v, keys, f.Name+".")
					continue
				default:
					continue
				}
			}

			keys[prefix+f.Name] = m
		}

		m.Record(v.Interface())
	}
}
