// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"encoding/json"
	"log"
	"math/rand"
	"testing"
	"time"
)

type SomeMetrics struct {
	Fail bool
}

type MoreMetrics struct {
	Request   bool
	Bytes     int
	Speed     float64
	Latency   time.Duration
	Labels    string
	Map       map[string]SomeMetrics
	Responses SomeMetrics
}

func TestMetric(t *testing.T) {
	m := Monitor{
		Name: "metric-test",
	}

	reports := make(chan *Summary)

	m.PublishFunc(func(s *Summary) {
		reports <- s
	})

	m.Start()

	m.RecordMetrics("test", &MoreMetrics{
		Request: true,
		Bytes:   123,
		Speed:   123.456,
		Latency: time.Millisecond * 2,
		Labels:  "asdf",
		Map: map[string]SomeMetrics{
			"a": SomeMetrics{
				Fail: true,
			},
			"b": SomeMetrics{
				Fail: false,
			},
		},
	})

	m.RecordMetrics("test", &MoreMetrics{
		Request: true,
		Bytes:   456,
		Speed:   456.789,
		Latency: time.Millisecond * 4,
		Labels:  "asdf",
		Map: map[string]SomeMetrics{
			"a": SomeMetrics{
				Fail: true,
			},
			"c": SomeMetrics{
				Fail: true,
			},
		},
		Responses: SomeMetrics{
			Fail: true,
		},
	})

	result := <-reports

	if len(result.Data) != 1 {
		t.Fatalf("expecting metrics %d\n", len(result.Data))
	}

	metrics, ok := result.Data["test"]
	if !ok {
		t.Fatalf("expecting metrics 'data' in map '%#v'\n", result.Data)
	}

	if len(metrics.Keys) != 7 {
		t.Fatalf("expecting 7 keys instead of %d\n", len(metrics.Keys))
	}

	values := map[string]string{
		"Request":        `{"Count":2}`,
		"Bytes":          `{"Average":289.5,"Maximum":456,"Minimum":123}`,
		"Speed":          `{"Average":290.1225,"Maximum":456.789,"Minimum":123.456}`,
		"Latency":        `{"Average":3e+06,"Maximum":4000000,"Minimum":2000000}`,
		"Labels":         `{"Items":{"asdf":2}}`,
		"Map":            `{"Items":{"a":{"Fail":{"Count":2}},"b":{"Fail":{"Count":0}},"c":{"Fail":{"Count":1}}}}`,
		"Responses.Fail": `{"Count":1}`,
	}

	for key, value := range values {
		text, err := json.Marshal(metrics.Keys[key])
		if err != nil {
			t.Fatalf("%s\n", err)
		}

		s := string(text)
		if s != value {
			t.Fatalf("result should be '%s' instead of '%s'\n", value, s)
		}
	}
}

func BenchmarkMetric(b *testing.B) {
	n := 10
	metrics := make([]*MoreMetrics, n)

	s := [...]string{"a", "b", "c", "d", "e", "f", "g", "h"}

	for i := 0; i != n; i++ {
		metrics[i] = &MoreMetrics{
			Request: rand.Int()%2 == 0,
			Bytes:   rand.Int() % 1024,
			Speed:   rand.Float64(),
			Latency: time.Duration(rand.Int() % 1000000),
			Labels:  s[rand.Int()%len(s)],
			Map: map[string]SomeMetrics{
				"a": SomeMetrics{
					Fail: rand.Int()%2 == 0,
				},
				"b": SomeMetrics{
					Fail: rand.Int()%2 == 0,
				},
			},
		}
	}

	m := Monitor{
		Name: "random-test",
	}

	output := false

	m.PublishFunc(func(s *Summary) {
		if output {
			text, err := json.Marshal(s)
			if err != nil {
				b.Fatal(err)
			}

			log.Printf("\n%s\n", string(text))
		}
	})

	m.Start()
	b.ResetTimer()

	k := 0

	for i := 0; i != b.N; i++ {
		metric := metrics[k]
		k = (k + 1) % n
		m.RecordMetrics("test", metric)
	}
}
