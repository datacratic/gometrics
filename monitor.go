// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Publisher defines an interface where the summary of metrics can be published.
type Publisher interface {
	Send(s *Summary)
}

// DefaultMetricPublishInterval defines the default frequency used to publish the summary of metrics.
var DefaultMetricPublishInterval = time.Second

// Monitor implements a service where metrics can be posted.
type Monitor struct {
	// Name contains an friendly identifier for this set of metrics.
	Name string
	// Publisher represents the handler that will receives the summary of metrics periodically.
	Publisher Publisher
	// PublishInterval contains the frequency of publication. Will use DefaultMetricPublishInterval if 0.
	PublishInterval time.Duration

	summary *Summary
	records chan interface{}
}

// Start creates the background service that aggregate metrics and publish them periodically.
func (monitor *Monitor) Start() (err error) {
	if monitor.Name == "" {
		err = fmt.Errorf("missing monitor name")
		return
	}

	if monitor.Publisher == nil {
		monitor.PublishFunc(func(s *Summary) {
			text, err := json.MarshalIndent(s, "", "\t")
			if err != nil {
				panic(err.Error())
			}

			fmt.Println(string(text))
		})
	}

	monitor.summary = &Summary{
		Name: monitor.Name,
		Keys: make(map[string]Metric),
	}

	if monitor.PublishInterval == 0 {
		monitor.PublishInterval = DefaultMetricPublishInterval
	}

	monitor.records = make(chan interface{})

	// start the background service
	go func() {
		t := time.NewTicker(monitor.PublishInterval)
		for {
			select {
			case r := <-monitor.records:
				monitor.summary.Record(r)
				monitor.summary.Hits++

			case <-t.C:
				if monitor.summary.Hits == 0 {
					break
				}

				s := monitor.summary
				s.Time = time.Now()
				s.Send++
				monitor.summary = &Summary{
					Name: s.Name,
					Keys: make(map[string]Metric),
					Send: s.Send,
				}

				monitor.Publisher.Send(s)
			}
		}
	}()

	return
}

// RecordMetrics posts a set of metrics to the monitor background service.
func (monitor *Monitor) RecordMetrics(value interface{}) {
	monitor.records <- value
}

// Publish sets the publish handler.
func (monitor *Monitor) Publish(handler Publisher) {
	monitor.Publisher = handler
}

// PublishFunc defines an helper to support the Publisher interface.
type PublishFunc func(*Summary)

// Send invokes the function literal with the summary.
func (f PublishFunc) Send(s *Summary) {
	f(s)
}

// PublishFunc is an helper that sets a function literal as the publish handler.
func (monitor *Monitor) PublishFunc(handler func(*Summary)) {
	monitor.Publish(PublishFunc(handler))
}

// NewJSONMonitor creates a publisher that POST the summary of metrics as a JSON object to the specified URL.
func NewJSONMonitor(name string, url string) *Monitor {
	m := &Monitor{
		Name: name,
	}

	m.PublishFunc(func(s *Summary) {
		text, err := json.Marshal(s)
		if err != nil {
			panic(err.Error())
		}

		r, err := http.Post(url, "application/json", bytes.NewReader(text))
		if err != nil {
			panic(err.Error())
		}

		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err.Error())
		}

		r.Body.Close()
	})

	return m
}
