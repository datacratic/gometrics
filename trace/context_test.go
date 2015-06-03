// Copyright (c) 2014 Datacratic. All rights reserved.

package trace

import (
	"runtime"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestContext(t *testing.T) {
	c := Start(context.Background(), "test", "*")

	// do something
	time.Sleep(time.Millisecond)

	Leave(c, "bye")
}

func TestContextAsync(t *testing.T) {
	c := Start(context.Background(), "test", "*")

	go func() {
		c := Enter(c, "go")

		// do something
		time.Sleep(time.Millisecond)

		Leave(c, "done")
	}()

	Leave(c, "bye")
}

func BenchmarkEnterLeave(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := Enter(context.Background(), "Begin")
		Leave(c, "End")

		// allow some recycling
		runtime.Gosched()
	}
}

func BenchmarkInnerBeginEnd(b *testing.B) {
	c := Enter(context.Background(), "Begin")

	for i := 0; i < b.N; i++ {
		inner := Enter(c, "Request")
		Leave(inner, "Done")
	}

	Leave(c, "End")
}

func BenchmarkCount(b *testing.B) {
	c := Enter(context.Background(), "Begin")

	for i := 0; i < b.N; i++ {
		Count(c, "Len", i)
	}

	Leave(c, "End")
}
