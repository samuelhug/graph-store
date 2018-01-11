package graph

import (
	"bytes"
	"encoding/gob"
	"errors"
)

type graphGob struct {
	inv      map[*Vertex]string
	Vertices map[string]interface{}
	Edges    map[string]map[string]int
}

// add a key - vertex pair to the graphGob
func (g graphGob) add(v *Vertex) {
	// set the key - vertex pair
	g.Vertices[v.key] = v.value

	g.Edges[v.key] = map[string]int{}

	// for each outgoing edge...
	for neighbor, weight := range v.outgoingEdges {
		// save the edge connection to the neighbor into the edges map
		g.Edges[v.key][neighbor.key] = weight
	}
}

// GobEncode encodes the graph into a []byte. With this method, graph implements the gob.GobEncoder interface.
func (g *Graph) GobEncode() ([]byte, error) {
	// build inverted map
	inv := map[*Vertex]string{}
	for key, v := range g.vertices {
		if _, ok := inv[v]; !ok {
			inv[v] = key
		}
	}

	gGob := graphGob{inv, map[string]interface{}{}, map[string]map[string]int{}}

	// add vertices and edges to gGob
	for _, v := range g.vertices {
		gGob.add(v)
	}

	// encode gGob
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(gGob)

	return buf.Bytes(), err
}

// GobDecode eecodes a []byte into the graph's vertices and edges. With this method, graph implements the gob.GobDecoder interface.
func (g *Graph) GobDecode(b []byte) (err error) {
	// decode into graphGob
	gGob := &graphGob{}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(gGob)
	if err != nil {
		return
	}

	// set the vertices
	for key, value := range gGob.Vertices {
		g.Set(key, value)
	}

	// connect the vertices
	for key, neighbors := range gGob.Edges {
		for otherKey, weight := range neighbors {
			if ok := g.Connect(key, otherKey, weight); !ok {
				return errors.New("invalid edge endpoints")
			}
		}
	}

	return
}
