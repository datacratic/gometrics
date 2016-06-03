// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"time"

	"golang.org/x/net/context"
)

// Handler defines the interface needed to process captured events.
// The 1st event is used as a root node.
// Each event are captured within a context indicated by their 'From' field.
type Handler interface {
	HandleTrace([]Event)
	Report(time.Duration)
	Close()
}

// handlerKey is a private type to find the current handler.
type handlerKey int

// SetHandler installs an handler in a new context.
func SetHandler(c context.Context, h Handler) context.Context {
	return context.WithValue(c, handlerKey(0), h)
}

// AddHandler installs an additional handler in a new context to the handler already installed.
func AddHandler(c context.Context, h Handler) context.Context {
	item, ok := c.Value(handlerKey(0)).(Handler)
	if !ok {
		return context.WithValue(c, handlerKey(0), item)
	}

	return context.WithValue(c, handlerKey(0), multiHandler{item, h})
}

type NilHandler struct {
}

func (handler *NilHandler) HandleTrace(events []Event) {
}

func (handler *NilHandler) Report(dt time.Duration) {
}

func (handler *NilHandler) Close() {
}

type chainHandler struct {
	Handler
	f func()
}

func (h *chainHandler) HandleTrace(events []Event) {
	h.Handler.HandleTrace(events)
	h.f()
}

func (h *chainHandler) Report(dt time.Duration) {
	h.Handler.Report(dt)
}

func (h *chainHandler) Close() {
	h.Handler.Close()
}

type multiHandler []Handler

func (items multiHandler) HandleTrace(events []Event) {
	for _, item := range items {
		item.HandleTrace(events)
	}
}

func (items multiHandler) Report(dt time.Duration) {
	for _, item := range items {
		item.Report(dt)
	}
}

func (items multiHandler) Close() {
	for _, item := range items {
		item.Close()
	}
}
