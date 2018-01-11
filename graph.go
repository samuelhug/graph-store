// Package graph implements a weighted, directed graph data structure.
// See https://en.wikipedia.org/wiki/Graph_(abstract_data_type) for more information.
package graph

import (
	"errors"
	"sync"
)

// Vertex reprsents a vertex in a graph
type Vertex struct {
	key           string
	value         interface{}     // the stored value
	incomingEdges map[*Vertex]int // maps the incoming edge to its weight
	outgoingEdges map[*Vertex]int // maps the outgoing edge to its weight
	sync.RWMutex
}

// GetIncoming returns the map of incoming edges and their weights.
func (v *Vertex) GetIncoming() map[*Vertex]int {
	if v == nil {
		return nil
	}

	v.RLock()
	incomingEdges := v.incomingEdges
	v.RUnlock()

	return incomingEdges
}

// GetOutgoing returns the map of outgoing edges and their weights.
func (v *Vertex) GetOutgoing() map[*Vertex]int {
	if v == nil {
		return nil
	}

	v.RLock()
	outgoingEdges := v.outgoingEdges
	v.RUnlock()

	return outgoingEdges
}

// Key returns the Vertex's key.
func (v *Vertex) Key() string {
	if v == nil {
		return ""
	}

	v.RLock()
	key := v.key
	v.RUnlock()

	return key
}

// Value returns the Vertex's value.
func (v *Vertex) Value() interface{} {
	if v == nil {
		return nil
	}

	v.RLock()
	value := v.value
	v.RUnlock()

	return value
}

// Graph reprsents a structure containing multiple interconnected vertices
type Graph struct {
	vertices map[string]*Vertex // A map of all the vertices in this graph, indexed by their key.
	sync.RWMutex
}

// New initializes a new graph.
func New() *Graph {
	return &Graph{map[string]*Vertex{}, sync.RWMutex{}}
}

// Len returns the number of vertices contained in the graph.
func (g *Graph) Len() int {
	return len(g.vertices)
}

// Set creates a new vertex and stores the given value if there is no vertex with the specified key yet.
// Otherwise, it updates the value, but leaves all connections intact.
func (g *Graph) Set(key string, value interface{}) {
	// lock graph until this method is finished to prevent changes made by other goroutines
	g.Lock()
	defer g.Unlock()

	v := g.get(key)

	// if no such node exists
	if v == nil {
		// create a new one
		v = &Vertex{key, value, map[*Vertex]int{}, map[*Vertex]int{}, sync.RWMutex{}}

		// and add it to the graph
		g.vertices[key] = v

		return
	}

	// else, just update the value
	v.Lock()
	v.value = value
	v.Unlock()
}

// Delete the vertex with the specified key. Return false if key is invalid.
func (g *Graph) Delete(key string) bool {
	// lock graph until this method is finished to prevent changes made by other goroutines while this one is looping etc.
	g.Lock()
	defer g.Unlock()

	// get vertex in question
	v := g.get(key)
	if v == nil {
		return false
	}

	// iterate over incomingEdges, remove edges from vertices
	for neighbor := range v.incomingEdges {
		// delete edge to the to-be-deleted vertex
		neighbor.Lock()
		delete(neighbor.outgoingEdges, v)
		neighbor.Unlock()
	}

	// iterate over outgoingEdges, remove edges from vertices
	for neighbor := range v.outgoingEdges {
		// delete edge to the to-be-deleted vertex
		neighbor.Lock()
		delete(neighbor.incomingEdges, v)
		neighbor.Unlock()
	}

	// delete vertex
	delete(g.vertices, key)

	return true
}

// GetAll returns a slice containing all vertices. The slice is empty if the graph contains no nodes.
func (g *Graph) GetAll() (all []*Vertex) {
	g.RLock()
	for _, v := range g.vertices {
		all = append(all, v)
	}
	g.RUnlock()

	return
}

// Get returns the vertex with this key, or nil and an error if there is no vertex with this key.
func (g *Graph) Get(key string) (v *Vertex, err error) {
	g.RLock()
	v = g.get(key)
	g.RUnlock()

	if v == nil {
		err = errors.New("graph: invalid key")
	}

	return
}

// get is an internal function, does NOT lock the graph, should only be used in between RLock() and RUnlock() (or Lock() and Unlock()).
func (g *Graph) get(key string) *Vertex {
	return g.vertices[key]
}

// Connect creates a directed edge between the vertices specified by fromKey and toKey. Returns false if one or both of the keys are invalid or if they are the same.
// If there already is a connection, it is overwritten with the new edge weight.
func (g *Graph) Connect(fromKey string, toKey string, weight int) bool {
	// recursive edges are forbidden
	if fromKey == toKey {
		return false
	}

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get vertices and check for validity of keys
	fromV := g.get(fromKey)
	toV := g.get(toKey)

	if fromV == nil || toV == nil {
		return false
	}

	// add connection to both vertices
	fromV.Lock()
	toV.Lock()

	fromV.outgoingEdges[toV] = weight
	toV.incomingEdges[fromV] = weight

	fromV.Unlock()
	toV.Unlock()

	// success
	return true
}

// Disconnect removes an edge connecting the two vertices. Returns false if one or both of the keys are invalid or if they are the same.
func (g *Graph) Disconnect(fromKey string, toKey string) bool {
	// recursive edges are forbidden
	if fromKey == toKey {
		return false
	}

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get vertices and check for validity of keys
	fromV := g.get(fromKey)
	toV := g.get(toKey)

	if fromV == nil || toV == nil {
		return false
	}

	// delete the edge from both vertices
	fromV.Lock()
	toV.Lock()

	delete(fromV.outgoingEdges, toV)
	delete(toV.incomingEdges, fromV)

	fromV.Unlock()
	toV.Unlock()

	return true
}

// IsConnected returns true and the edge weight if there is an edge from fromKey to toKey.
// Returns false if one or both keys are invalid, if they are the same, or if there is no edge connecting them.
func (g *Graph) IsConnected(fromKey string, toKey string) (exists bool, weight int) {
	// sanity check
	if fromKey == toKey {
		return
	}

	g.RLock()

	fromV := g.get(fromKey)
	if fromV == nil {
		g.RUnlock()
		return
	}

	toV := g.get(toKey)
	if toV == nil {
		g.RUnlock()
		return
	}

	g.RUnlock()

	fromV.RLock()
	defer fromV.RUnlock()
	toV.RLock()
	defer toV.RUnlock()

	// choose vertex with less edges (easier to find 1 in 10 than to find 1 in 100)
	if len(fromV.outgoingEdges) < len(toV.incomingEdges) {
		// iterate over it's map of edges; when the right vertex is found, return
		for iteratingV, weight := range fromV.outgoingEdges {
			if iteratingV == toV {
				return true, weight
			}
		}
	} else {
		// iterate over it's map of edges; when the right vertex is found, return
		for iteratingV, weight := range toV.incomingEdges {
			if iteratingV == fromV {
				return true, weight
			}
		}
	}

	return
}
