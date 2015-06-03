// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Logs generates a log line for each metric.
type Logs struct {
	// Filename is optional and contains the name of the log file.
	// When empty, logs will use stderr.
	// If the filename contains a '%s', it will be replace by the UTC time in RFC3339 format.
	Filename string
	// Prefix contains the prefix that appears at the beginning of each generated log line.
	Prefix string

	logger *log.Logger
	writer *bufio.Writer
	once   sync.Once
}

// NewWriter returns a writer that will writes all metrics to its logger.
func (logs *Logs) NewWriter(s *Summary) Writer {
	logs.once.Do(logs.initialize)

	return &logWriter{
		logs: logs, dt: s.Step.Seconds(),
	}
}

func (logs *Logs) initialize() {
	fd := os.Stderr

	path := logs.Filename
	if path != "" {
		path = strings.Replace(path, "%s", time.Now().UTC().Format(time.RFC3339), 1)

		var err error
		fd, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("failed to create log file '%s': %s\n", path, err)
		}
	}

	logs.writer = bufio.NewWriter(fd)
	logs.logger = log.New(logs.writer, logs.Prefix, log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

type logWriter struct {
	logs *Logs
	dt   float64
}

func (w *logWriter) Write(name string, value float64) (err error) {
	w.logs.logger.Printf("%s=%f\n", name, value)
	return
}

func (w *logWriter) WriteScaled(name string, value float64) (err error) {
	w.logs.logger.Printf("%s=%f\n", name, value/w.dt)
	return
}

func (w *logWriter) WriteString(name, text string) (err error) {
	w.logs.logger.Printf("%s=%s\n", name, text)
	return
}

func (w *logWriter) Close() {
	w.logs.writer.Flush()
}
