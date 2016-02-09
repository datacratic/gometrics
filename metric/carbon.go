// Copyright (c) 2015 Datacratic. All rights reserved.

package metric

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Carbon enables writing summary of metrics to Carbon daemons at the specified URLs.
type Carbon struct {
	// URLs contains a list of addresses used to dial e.g. tcp://127.0.0.1:2023.
	URLs []string
	// Prefix contains the path under which all keys will be written.
	Prefix string

	once sync.Once
	conn []*carbonConn
	path string
}

// NewWriter creates a new Carbon writer that will send the aggregated summary metrics to all connections.
// In case of failure, the faulty connection is closed and the written data is queued.
// Metrics can still be sent while the connection is reestablished.
// This background process happens in the background and will retry indefinitely.
func (carbon *Carbon) NewWriter(s *Summary) (result Writer) {
	carbon.once.Do(carbon.initialize)

	return &carbonWriter{
		carbon: carbon,
		dt:     s.Step.Seconds(),
		format: fmt.Sprintf("%s%%s %%f %d\n", carbon.path, s.Time.Unix()),
	}
}

func (carbon *Carbon) initialize() {
	path := carbon.Prefix
	if path != "" && !strings.HasSuffix(path, ".") {
		path += "."
	}

	carbon.path = path
	carbon.conn = make([]*carbonConn, len(carbon.URLs))

	for i, url := range carbon.URLs {
		conn := carbon.connect(i, url)
		go func() {
			for f := range conn.feed {
				f()
			}
		}()
	}
}

func (carbon *Carbon) connect(i int, address string) *carbonConn {
	u, err := url.Parse(address)
	if err != nil {
		log.Fatalf("url '%s': %s", address, err)
	}

	conn := &carbonConn{
		feed:    make(chan func()),
		network: u.Scheme,
		address: u.Host,
	}

	carbon.conn[i] = conn
	return conn
}

type carbonWriter struct {
	carbon *Carbon
	buffer bytes.Buffer
	dt     float64
	prefix string
	format string
}

func (w *carbonWriter) Write(name string, value float64) (err error) {
	_, err = fmt.Fprintf(&w.buffer, w.format, name, value)
	return
}

func (w *carbonWriter) WriteScaled(name string, value float64) (err error) {
	_, err = fmt.Fprintf(&w.buffer, w.format, name, value/w.dt)
	return
}

func (w *carbonWriter) WriteString(name, text string) (err error) {
	err = ErrIgnored
	return
}

func (w *carbonWriter) Close() {
	go func() {
		var wg sync.WaitGroup

		for i, n := 0, len(w.carbon.conn); i < n; i++ {
			wg.Add(1)
			conn := w.carbon.conn[i]
			conn.feed <- func() {
				conn.send(w.buffer.Bytes())
				wg.Done()
			}
		}

		wg.Wait()
	}()
}

type carbonConn struct {
	conn    net.Conn
	feed    chan func()
	network string
	address string
}

func (carbon *carbonConn) send(data []byte) {
	retry := 0
	sleep := time.Second

	for {
		if retry != 0 {
			log.Printf("carbon: connect attempt %d to '%s://%s'\n", retry, carbon.network, carbon.address)
		}

		err := carbon.write(data)
		if err == nil {
			break
		}

		log.Printf("carbon: %s\n", err)

		if carbon.conn != nil {
			if err = carbon.conn.Close(); err != nil {
				log.Printf("carbon: %s\n", err)
			}

			carbon.conn = nil
		}

		time.Sleep(sleep)
		if sleep < time.Minute {
			sleep += sleep
		}

		retry++
	}
}

func (carbon *carbonConn) write(data []byte) (err error) {
	if carbon.conn == nil {
		carbon.conn, err = net.Dial(carbon.network, carbon.address)
		if err != nil {
			return
		}

		log.Printf("carbon: connected at '%s://%s'\n", carbon.network, carbon.address)
	}

	io.Copy(os.Stderr, bytes.NewReader(data))
	_, err = io.Copy(carbon.conn, bytes.NewReader(data))
	return
}
