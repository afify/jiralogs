package csync

import "sync"

// Slice provides thread-safe slice operations
type Slice[T any] struct {
	mu    sync.RWMutex
	items []T
}

// NewSlice creates a new thread-safe slice
func NewSlice[T any]() *Slice[T] {
	return &Slice[T]{
		items: make([]T, 0),
	}
}

// Append adds an item to the slice
func (s *Slice[T]) Append(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// Seq returns an iterator over the slice
func (s *Slice[T]) Seq() func(func(T) bool) {
	s.mu.RLock()
	items := make([]T, len(s.items))
	copy(items, s.items)
	s.mu.RUnlock()
	
	return func(yield func(T) bool) {
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}