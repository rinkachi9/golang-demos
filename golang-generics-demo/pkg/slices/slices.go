package slices

// Map transforms a slice of T to a slice of R.
func Map[T any, R any](items []T, transform func(T) R) []R {
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}
	return result
}

// Filter returns a slice of items that satisfy the predicate.
func Filter[T any](items []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Reduce accumulates a value from the slice.
func Reduce[T any, R any](items []T, initial R, reducer func(R, T) R) R {
	acc := initial
	for _, item := range items {
		acc = reducer(acc, item)
	}
	return acc
}

// Chunk splits a slice into chunks of the given size.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Unique returns a new slice removing duplicate values using a map for tracking.
// Note: T must be comparable.
func Unique[T comparable](items []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Contains checks if an item exists in the slice.
func Contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
