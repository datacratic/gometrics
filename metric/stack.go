// Copyright (c) 2014 Datacratic. All rights reserved.

package metric

// Stack implements a fallback mechanism when a writer ignores a type of metric.
// For each specified reporter, an associated writer is created.
// When a writer indicates that the value was ignored, the next writer is used until the stack is exhausted.
type Stack struct {
	// Items contains the ordered list of reporters.
	Items []Reporter
}

func NewStack(items ...Reporter) Reporter {
	return &Stack{Items: items}
}

// NewWriter returns the stack of writers that will be used to write metrics.
func (stack *Stack) NewWriter(s *Summary) (result Writer) {
	w := new(stackWriter)
	for _, item := range stack.Items {
		w.list = append(w.list, item.NewWriter(s))
	}

	return w
}

type stackWriter struct {
	list []Writer
}

func (stack *stackWriter) Write(name string, value float64) (err error) {
	for _, w := range stack.list {
		err = w.Write(name, value)
		if err != ErrIgnored {
			break
		}
	}

	return
}

func (stack *stackWriter) WriteScaled(name string, value float64) (err error) {
	for _, w := range stack.list {
		err = w.WriteScaled(name, value)
		if err != ErrIgnored {
			break
		}
	}

	return
}

func (stack *stackWriter) WriteString(name, text string) (err error) {
	for _, w := range stack.list {
		err = w.WriteString(name, text)
		if err != ErrIgnored {
			break
		}
	}

	return
}

func (stack *stackWriter) Close() {
	for _, w := range stack.list {
		w.Close()
	}
}
