// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import "time"

func ExampleMonitor() {
	type SomeMetrics struct {
		Request bool
		Bytes   int
		Latency time.Duration
	}

	d := time.Millisecond * 100

	monitor := Monitor{
		Name:            "my metrics",
		PublishInterval: d,
	}

	// start the monitoring background service
	monitor.Start()

	monitor.RecordMetrics("test", &SomeMetrics{
		Request: true,
		Bytes:   100,
		Latency: time.Millisecond * 2,
	})

	monitor.RecordMetrics("test", &SomeMetrics{
		Request: true,
		Bytes:   150,
		Latency: time.Millisecond * 4,
	})

	// wait for the next publication
	time.Sleep(d * 2)
}
