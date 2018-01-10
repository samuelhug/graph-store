// Package graph implements a weighted, undirected graph data structure.
// See https://en.wikipedia.org/wiki/Graph_(abstract_data_type) for more information.
package graph

import (
	"errors"
	"sync"
)

// Vertex reprsents a vertex in a graph
type Vertex struct {
	key       string
	value     interface{}     // the stored value
	neighbors map[*Vertex]int // maps the neighbor node to the weight of the connection to it
	sync.RWMutex
}

// GetNeighbors returns the map of neighbors.
func (v *Vertex) GetNeighbors() map[*Vertex]int {
	if v == nil {
		return nil
	}

	v.RLock()
	neighbors := v.neighbors
	v.RUnlock()

	return neighbors
}

// Key returns the vertexes key.
func (v *Vertex) Key() string {
	if v == nil {
		return ""
	}

	v.RLock()
	key := v.key
	v.RUnlock()

	return key
}

// Value returns the Vertexes value.
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
	vertexes map[string]*Vertex // A map of all the vertexes in this graph, indexed by their key.
	sync.RWMutex
}

// New initializes a new graph.
func New() *Graph {
	return &Graph{map[string]*Vertex{}, sync.RWMutex{}}
}

// Len returns the amount of vertexes contained in the graph.
func (g *Graph) Len() int {
	return len(g.vertexes)
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
		v = &Vertex{key, value, map[*Vertex]int{}, sync.RWMutex{}}

		// and add it to the graph
		g.vertexes[key] = v

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

	// iterate over neighbors, remove edges from neighboring vertexes
	for neighbor := range v.neighbors {
		// delete edge to the to-be-deleted vertex
		neighbor.Lock()
		delete(neighbor.neighbors, v)
		neighbor.Unlock()
	}

	// delete vertex
	delete(g.vertexes, key)

	return true
}

// GetAll returns a slice containing all vertexes. The slice is empty if the graph contains no nodes.
func (g *Graph) GetAll() (all []*Vertex) {
	g.RLock()
	for _, v := range g.vertexes {
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
	return g.vertexes[key]
}

// Connect creates an edge between the vertexes specified by the keys. Returns false if one or both of the keys are invalid or if they are the same.
// If there already is a connection, it is overwritten with the new edge weight.
func (g *Graph) Connect(key string, otherKey string, weight int) bool {
	// recursive edges are forbidden
	if key == otherKey {
		return false
	}

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get vertexes and check for validity of keys
	v := g.get(key)
	otherV := g.get(otherKey)

	if v == nil || otherV == nil {
		return false
	}

	// add connection to both vertexes
	v.Lock()
	otherV.Lock()

	v.neighbors[otherV] = weight
	otherV.neighbors[v] = weight

	v.Unlock()
	otherV.Unlock()

	// success
	return true
}

// Disconnect removes an edge connecting the two vertexes. Returns false if one or both of the keys are invalid or if they are the same.
func (g *Graph) Disconnect(key string, otherKey string) bool {
	// recursive edges are forbidden
	if key == otherKey {
		return false
	}

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get vertexes and check for validity of keys
	v := g.get(key)
	otherV := g.get(otherKey)

	if v == nil || otherV == nil {
		return false
	}

	// delete the edge from both vertexes
	v.Lock()
	otherV.Lock()

	delete(v.neighbors, otherV)
	delete(otherV.neighbors, v)

	v.Unlock()
	otherV.Unlock()

	return true
}

// Adjacent returns true and the edge weight if there is an edge between the vertexes specified by their keys.
// Returns false if one or both keys are invalid, if they are the same, or if there is no edge between the vertexes.
func (g *Graph) Adjacent(key string, otherKey string) (exists bool, weight int) {
	// sanity check
	if key == otherKey {
		return
	}

	g.RLock()

	v := g.get(key)
	if v == nil {
		g.RUnlock()
		return
	}

	otherV := g.get(otherKey)
	if otherV == nil {
		g.RUnlock()
		return
	}

	g.RUnlock()

	v.RLock()
	defer v.RUnlock()
	otherV.RLock()
	defer otherV.RUnlock()

	// choose vertex with less edges (easier to find 1 in 10 than to find 1 in 100)
	if len(v.neighbors) < len(otherV.neighbors) {
		// iterate over it's map of edges; when the right vertex is found, return
		for iteratingV, weight := range v.neighbors {
			if iteratingV == otherV {
				return true, weight
			}
		}
	} else {
		// iterate over it's map of edges; when the right vertex is found, return
		for iteratingV, weight := range otherV.neighbors {
			if iteratingV == v {
				return true, weight
			}
		}
	}

	return
}
