// Copyright (c) 2015 Datacratic. All rights reserved.

package metric

import (
	"fmt"
	"log"
	"strings"
)

// Labels counts and tracks generated log lines.
type Labels struct {
	lines map[string]int
	count int
}

// Record tracks generated lines of text for the specified label.
func (labels *Labels) Record(value interface{}) {
	text := ""

	switch item := value.(type) {
	case []string:
		text = strings.Join(item, " ")
	case string:
		text = item
	default:
		text = fmt.Sprintf("%v", item)
	}

	if text == "" {
		log.Printf("cannot record an empty string")
		return
	}

	if labels.lines == nil {
		labels.lines = make(map[string]int)
	}

	labels.lines[text]++
	labels.count++
}

// Reset clears the label and set its count to zero.
func (labels *Labels) Reset() {
	labels.lines = nil
	labels.count = 0
}

// Write creates a key for the label that represents the number entries scaled over time.
// Each unique lines of text will also generate a text key.
// If needed, the count of an individual line is added in parenthesis at the end of the line.
func (labels *Labels) Write(w Writer, name string) {
	if labels.lines == nil {
		return
	}

	for text, n := range labels.lines {
		if n > 1 {
			w.WriteString(name, fmt.Sprintf("%s (%d)", text, n))
		} else {
			w.WriteString(name, text)
		}
	}

	w.WriteScaled(name, float64(labels.count))
}
