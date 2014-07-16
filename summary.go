// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"reflect"
	"time"
)

// Summary represents the aggregation of metrics over a period of time.
type Summary struct {
	// Name contains a friendly identifier for this set of metrics.
	Name string
	// Keys contains the current set of metrics.
	Keys map[string]Metric
	// Hits contains the number of accumulated samples.
	Hits int
	// Send contains a sequential number of that summary based on the number of times it was published by the monitor.
	Send int
	// Time contains the time stamp of the summary publication.
	Time time.Time
}

// Record uses reflection to create and aggregate metrics based on their data type.
func (summary *Summary) Record(value interface{}) {
	m := reflect.ValueOf(value).Elem()
	recordMembers(m, summary.Keys)
}

func recordMembers(m reflect.Value, keys map[string]Metric) {
	t := m.Type()
	for i := 0; i < m.NumField(); i++ {
		f := m.Field(i)
		n := t.Field(i).Name

		k, ok := keys[n]
		if !ok {
			switch f.Interface().(type) {
			case bool:
				k = new(MetricBool)
			case int:
				k = new(MetricInt)
			case float64:
				k = new(MetricFloat)
			case time.Duration:
				k = new(MetricDuration)
			case string:
				k = &MetricString{
					Items: make(map[string]int),
				}
			default:
				switch {
				case f.Type().Kind() == reflect.Map:
					k = &MetricMap{
						Items: make(map[string]map[string]Metric),
					}
				default:
					continue
				}
			}

			keys[n] = k
		}

		k.Record(f.Interface())
	}
}
