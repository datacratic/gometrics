// Copyright (c) 2015 Datacratic. All rights reserved.

package metric

import (
	"fmt"
	"io"

	"os"
	"time"
)

type Console struct {
	io.Writer
}

func (c *Console) NewWriter(s *Summary) Writer {
	w := c.Writer
	if nil == w {
		w = os.Stdout
	}

	t := s.Time.Format(time.RFC3339)
	return &consoleWriter{
		w: w,
		t: t,
		f: fmt.Sprintf("%s %%s %%f\n", t),
		d: s.Step.Seconds(),
	}
}

type consoleWriter struct {
	logs map[string]map[string]int64

	w io.Writer
	t string
	f string
	d float64
}

func (w *consoleWriter) Write(name string, value float64) (err error) {
	_, err = fmt.Fprintf(w.w, w.f, name, value)
	return
}

func (w *consoleWriter) WriteScaled(name string, value float64) (err error) {
	_, err = fmt.Fprintf(w.w, w.f, name, value/w.d)
	return
}

func (w *consoleWriter) WriteString(name, text string) (err error) {
	c, ok := w.logs[name]
	if !ok {
		c = make(map[string]int64)
		if nil == w.logs {
			w.logs = make(map[string]map[string]int64)
		}

		w.logs[name] = c
	}

	c[text]++
	return
}

func (w *consoleWriter) Close() {
	for name, items := range w.logs {
		for text, count := range items {
			if count == 1 {
				fmt.Fprintf(w.w, "%s %s %s\n", w.t, name, text)
			} else {
				fmt.Fprintf(w.w, "%s %s %s (%d)\n", w.t, name, text, count)
			}
		}
	}
}
