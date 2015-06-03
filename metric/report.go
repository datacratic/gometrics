// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import "errors"

var (
	// ErrIgnored indicates that a writer ignores a metric.
	ErrIgnored = errors.New("ignored")
)

// Writer implements a kind of reporting for every metrics.
type Writer interface {
	Write(name string, value float64) error
	WriteScaled(name string, value float64) error
	WriteString(name, text string) error
	Close()
}

// Reporter defines a kind of reporting for a summary of metrics.
type Reporter interface {
	NewWriter(s *Summary) Writer
}
