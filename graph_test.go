package graph

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	g := New()

	// set some vertices
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// make some connections
	ok := g.Connect("1", "2", 5)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("2", "3", 1)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("3", "1", 9)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("4", "2", 3)
	if !ok {
		t.Fail()
	}

	// test connections
	ok, weight := g.IsConnected("1", "2")
	if !ok || weight != 5 {
		t.Fail()
	}

	ok, weight = g.IsConnected("2", "3")
	if !ok || weight != 1 {
		t.Fail()
	}

	ok, weight = g.IsConnected("3", "1")
	if !ok || weight != 9 {
		t.Fail()
	}

	ok, weight = g.IsConnected("4", "2")
	if !ok || weight != 3 {
		t.Fail()
	}

	// test connections in the reverse (shouldn't work)
	ok, _ = g.IsConnected("2", "1")
	if ok {
		t.Fail()
	}

	ok, _ = g.IsConnected("3", "3")
	if ok {
		t.Fail()
	}

	ok, _ = g.IsConnected("1", "3")
	if ok {
		t.Fail()
	}

	ok, _ = g.IsConnected("2", "4")
	if ok {
		t.Fail()
	}

	// test non-connections
	ok, _ = g.IsConnected("1", "4")
	if ok {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	g := New()

	// set some vertices
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// make some connections
	ok := g.Connect("1", "2", 5)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("2", "3", 1)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("3", "1", 9)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("4", "2", 3)
	if !ok {
		t.Fail()
	}

	// preserve a pointer to node "1"
	one := g.get("1")
	if one == nil {
		t.Fail()
	}

	// delete node
	ok = g.Delete("1")
	if !ok {
		t.Fail()
	}

	// make sure it's not in the graph anymore
	deletedOne := g.get("1")
	if deletedOne != nil {
		t.Fail()
	}

	// test for orphaned connections
	neighbors := g.get("2").GetIncoming()
	for n := range neighbors {
		if n == one {
			t.Fail()
		}
	}

	neighbors = g.get("3").GetOutgoing()
	for n := range neighbors {
		if n == one {
			t.Fail()
		}
	}
}

func TestGob(t *testing.T) {
	g := New()

	// set key → value pairs
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// connect vertices/nodes
	g.Connect("1", "2", 5)
	g.Connect("1", "3", 1)
	g.Connect("2", "3", 9)
	g.Connect("4", "2", 3)

	// encode
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	err := enc.Encode(g)
	if err != nil {
		fmt.Println(err)
	}

	// now decode into new graph
	dec := gob.NewDecoder(buf)
	newG := New()
	err = dec.Decode(newG)
	if err != nil {
		fmt.Println(err)
	}

	// validate length of new graph
	if len(g.vertices) != len(newG.vertices) {
		t.Fail()
	}

	// validate contents of new graph
	for k, v := range g.vertices {
		if newV := newG.get(k); newV.value != v.value {
			t.Fail()
		}
	}
}

func ExampleGraph() {
	g := New()

	// set key → value pairs
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// connect vertices/nodes
	g.Connect("1", "2", 5)
	g.Connect("2", "3", 1)
	g.Connect("3", "1", 9)
	g.Connect("4", "2", 3)

	// delete a node, and all connections to it
	g.Delete("1")

	// encode into buffer
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	err := enc.Encode(g)
	if err != nil {
		fmt.Println(err)
	}

	// now decode into new graph
	dec := gob.NewDecoder(buf)
	newG := New()
	err = dec.Decode(newG)
	if err != nil {
		fmt.Println(err)
	}
}

func printVertices(vSlice map[string]*Vertex) {
	for _, v := range vSlice {
		fmt.Printf("%v\n", v.value)
		for otherV := range v.outgoingEdges {
			fmt.Printf("  → %v\n", otherV.value)
		}
		for otherV := range v.incomingEdges {
			fmt.Printf("  ← %v\n", otherV.value)
		}
	}
}
