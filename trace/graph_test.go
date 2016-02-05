// Copyright (c) 2014 Datacratic. All rights reserved.

package trace

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/context"
)

func TestGraph(t *testing.T) {
	g := &Graph{}
	c := SetHandler(context.Background(), g)
	c = Enter(c, "Begin")
	foo(c, "Hello")
	foo(c, "World")
	Leave(c, "End")

	svg, err := g.DrawSVG()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(svg))

	if len(svg) == 0 {
		t.Fatalf("missing svg")
	}

}

func foo(c context.Context, data string) {
	c = Start(c, "Foo", "*")

	Count(c, "Bytes", len(data))
	if !strings.Contains(data, "Hello") {
		bar(c)
		Leave(c, "Fail")
		return
	}

	Leave(c, "Done")
}

func bar(c context.Context) {
	c = Enter(c, "Bar")
	Leave(c, "Done")
}
