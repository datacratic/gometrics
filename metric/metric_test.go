// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	s := &Summary{
		Step: time.Second,
	}

	s.Count("c", true)
	s.Count("c", 10)
	s.Count("c", 0.5)

	s.Set("g", true)
	s.Set("g", -1)
	s.Set("g", 10)

	for i := 0; i < 100; i++ {
		s.Record("h", i)
	}

	s.Log("s", "hello")
	s.Log("s", "world")
	s.Log("s", "world")

	s.Write(&Logs{})
}
