// Copyright (c) 2014 Datacratic. All rights reserved.

/*
Package metric provides a simple way to accumulate and report metrics.

First, create a Monitor instance:

	m := Monitor{
		Name: "my metrics",
	}

	m.Start()

To record a set of metrics, create a custom structure where each member
represents one metric. If the type is supported, it will be recorded and
aggregated.

	type SomeMetrics struct {
		Request bool
		Bytes   int
		Latency time.Duration
	}

Supported types are currently limited to some basic types (bool, int, float64,
string, time.Duration) and can contain a map of string to another structure of
metrics.

	type LotsOfMetrics struct {
		InFlight int
		Workers  map[string]SomeMetrics
	}

A boolean metric is normally used for events as it counts the number of times it
was recorded as true. Numerical metrics (duration included) keep track of more
information like their minimal, maximum and average values. Finally, the string
metric counts the number of occurences of each recorded values.

*/
package metric
