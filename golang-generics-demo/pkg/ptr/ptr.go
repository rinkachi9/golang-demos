package ptr

// Of returns a pointer to the given value.
func Of[T any](v T) *T {
	return &v
}

// ValueOrDefault returns the value of the pointer if not nil, otherwise the default value.
func ValueOrDefault[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// Unwrap returns the value or the zero value if nil.
func Unwrap[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
