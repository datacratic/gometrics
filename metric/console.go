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
		f: fmt.Sprintf("%s %%s %%f\n", t),
		s: fmt.Sprintf("%s %%s %%s\n", t),
		d: s.Step.Seconds(),
	}
}

type consoleWriter struct {
	w io.Writer
	f string
	s string
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
	_, err = fmt.Fprintf(w.w, w.s, name, text)
	return
}

func (w *consoleWriter) Close() {
}
