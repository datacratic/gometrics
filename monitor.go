// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Publisher defines an interface where the summary of metrics can be published.
type Publisher interface {
	Send(s *Summary)
}

// PublisherFunc defines a convenience type to wrap a Publisher function.
type PublisherFunc func(s *Summary)

// Send sends the summary to the publisher functions.
func (fn PublisherFunc) Send(s *Summary) {
	fn(s)
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
	records chan metricSet
}

type metricSet struct {
	name string
	data interface{}
}

func (monitor *Monitor) Run() {
	monitor.Start()
}

// Start creates the background service that aggregate metrics and publish them periodically.
func (monitor *Monitor) Start() {
	if monitor.Name == "" {
		log.Panic("Name must be set for monitor")
	}

	if monitor.Publisher == nil {
		monitor.PublishFunc(func(s *Summary) {
			text, err := json.MarshalIndent(s, "", "\t")
			if err != nil {
				log.Panic(err)
			}

			log.Println(string(text))
		})
	}

	monitor.summary = &Summary{
		Name: monitor.Name,
		Data: make(map[string]*Metrics),
	}

	if monitor.PublishInterval == 0 {
		monitor.PublishInterval = DefaultMetricPublishInterval
	}

	monitor.records = make(chan metricSet)

	// start the background service
	go func() {
		t := time.NewTicker(monitor.PublishInterval)
		for {
			select {
			case r := <-monitor.records:
				monitor.summary.Record(r.name, r.data)

			case <-t.C:
				if len(monitor.summary.Data) == 0 {
					break
				}

				s := monitor.summary
				s.Time = time.Now()
				s.Step = monitor.PublishInterval
				s.Send++
				monitor.summary = &Summary{
					Name: s.Name,
					Data: make(map[string]*Metrics),
					Send: s.Send,
				}

				monitor.Publisher.Send(s)
			}
		}
	}()

	return
}

// RecordMetrics posts a set of metrics to the monitor background service.
func (monitor *Monitor) RecordMetrics(name string, data interface{}) {
	monitor.records <- metricSet{
		name: name,
		data: data,
	}
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

// NewJSONMonitor creates and starts a monitor that POST the summary of metrics as a JSON object to the specified URL.
func NewJSONMonitor(name string, url string) (monitor *Monitor) {
	monitor = &Monitor{
		Name: name,
	}

	monitor.PublishFunc(func(s *Summary) {
		text, err := json.Marshal(s)
		if err != nil {
			log.Panic(err)
		}

		r, err := http.Post(url, "application/json", bytes.NewReader(text))
		if err != nil {
			log.Panic(err)
		}

		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Panic(err)
		}

		r.Body.Close()
	})

	monitor.Start()
	return
}

var nullMonitor *Monitor

// NullMonitor defines a monitor that doesn't publish its results.
func NullMonitor() *Monitor {
	return nullMonitor
}

func init() {
	nullMonitor = &Monitor{
		Name:            "null",
		Publisher:       PublisherFunc(func(s *Summary) {}),
		PublishInterval: time.Second,
	}
	nullMonitor.Start()
}
