// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type CarbonMonitor struct {
	Name string
	URLs []string

	feed []chan *Summary
}

// NewCarbonMonitor creates a monitor that writes the summary of metrics to Carbon daemons at the specified URLs.
func NewCarbonMonitor(name string, urls []string) (monitor *Monitor) {
	monitor = &Monitor{
		Name: name,
	}

	feeds := make([]chan *Summary, len(urls))

	for i := range urls {
		feed := make(chan *Summary)
		feeds[i] = feed

		go func(url string) {
			wait := time.Second
			next := wait

			var conn net.Conn
			for {
				select {
				case <-time.After(next):
					if conn != nil {
						break
					}

					var err error
					conn, err = net.Dial("tcp", url)
					if err != nil {
						log.Printf("%s (retry in %v)\n", err, next)
						if next < time.Minute {
							next *= 2
						}

						break
					}

					log.Printf("connected to '%s'\n", url)
					next = wait
				case summary := <-feed:
					if conn == nil {
						break
					}

					when := summary.Time.Unix()
					dt := float64(summary.Step) / float64(time.Second)

					writer := bufio.NewWriter(conn)
					for name, metrics := range summary.Data {
						text := summary.Name + "." + name
						if err := writeToCarbon(writer, metrics.Keys, text, when, dt); err != nil {
							conn.Close()
							break
						}
					}

					writer.Flush()
				}
			}
		}(urls[i])
	}

	monitor.PublishFunc(func(s *Summary) {
		for i := range feeds {
			feeds[i] <- s
		}
	})

	return
}

func writeToCarbon(w io.Writer, metrics map[string]Metric, name string, when int64, dt float64) (err error) {
	write := func(name string, value float64) bool {
		_, err = fmt.Fprintf(w, "%s %f %d\n", name, value, when)
		return err != nil
	}

	for key, metric := range metrics {
		text := name + "." + key
		switch m := metric.(interface{}).(type) {
		case *Bool:
			write(text, float64(m.Count)/dt)
		case *Int:
			_ = write(text+".Average", m.Average) &&
				write(text+".Maximum", float64(m.Maximum)) &&
				write(text+".Minimum", float64(m.Minimum))
		case *Float:
			_ = write(text+".Average", m.Average) &&
				write(text+".Maximum", m.Maximum) &&
				write(text+".Minimum", m.Minimum)
		case *Duration:
			s := float64(time.Second)
			_ = write(text+".Average", m.Average/s) &&
				write(text+".Maximum", float64(m.Maximum)/s) &&
				write(text+".Minimum", float64(m.Minimum)/s)
		case *String:
			for key, value := range m.Items {
				if !write(text+"."+key, float64(value)/dt) {
					return
				}
			}
		case *Map:
			for key, value := range m.Items {
				if err = writeToCarbon(w, value, text+"."+key, when, dt); err != nil {
					return
				}
			}
		default:
			log.Panicf("unknown metric %+v", m)
		}
	}

	return
}
