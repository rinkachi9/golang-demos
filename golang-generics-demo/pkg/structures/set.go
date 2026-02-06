package structures

// Set is a generic set implementation using a map.
type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{data: make(map[T]struct{})}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

func (s *Set[T]) Add(item T) {
	s.data[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.data, item)
}

func (s *Set[T]) Contains(item T) bool {
	_, exists := s.data[item]
	return exists
}

func (s *Set[T]) Slice() []T {
	result := make([]T, 0, len(s.data))
	for item := range s.data {
		result = append(result, item)
	}
	return result
}
