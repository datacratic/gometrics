// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"flag"
	"strings"
	"time"

	"github.com/datacratic/gometrics/metric"
)

var defaultPeriod *time.Duration
var defaultCarbon *string

func init() {
	defaultPeriod = flag.Duration("metrics-period", 10*time.Second, "metrics reporting period")
	defaultCarbon = flag.String("metrics-carbon", "tcp://127.0.0.1:2003", "address to carbon endpoint(s)")
}

func New() Handler {
	return &Periodic{
		Period: *defaultPeriod,
		Handler: &Metrics{
			Prefix: "",
			Reporter: metric.NewStack(
				&metric.Carbon{
					URLs:   strings.Split(*defaultCarbon, ","),
					Prefix: "",
				},
				&metric.Console{},
			),
		},
	}

}
