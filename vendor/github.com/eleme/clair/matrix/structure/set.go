package structure

import "sync"

// Set implements the basic data structure set
// Note: it's not thread-safe
type Set struct {
	mu       sync.Mutex
	elements map[interface{}]bool
}

// NewSet returns an empty set
func NewSet() *Set {
	return &Set{elements: make(map[interface{}]bool), mu: sync.Mutex{}}
}

// Add adds an element into the set
func (set *Set) Add(element interface{}) {
	if !set.Contains(element) {
		set.mu.Lock()
		defer set.mu.Unlock()
		set.elements[element] = true

	}
}

// Rem removes an element into the set
func (set *Set) Rem(element interface{}) {
	if set.Contains(element) {
		set.mu.Lock()
		defer set.mu.Unlock()
		set.elements[element] = false

	}
}

// All returns all elements in the set
func (set *Set) All() []interface{} {
	var result []interface{}
	for k, v := range set.elements {
		if v {
			result = append(result, k)
		}
	}
	return result
}

// Contains check whether the set has the element
func (set *Set) Contains(element interface{}) bool {
	val, ok := set.elements[element]
	if val && ok {
		return true
	}
	return false
}

// Union unions sets
func Union(sets ...*Set) *Set {
	resultSet := NewSet()
	for _, set := range sets {
		if set != nil {
			for _, element := range set.All() {
				resultSet.Add(element)
			}
		}
	}
	return resultSet
}
