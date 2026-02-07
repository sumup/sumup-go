package ptr

// OrNil dereferences value and returns nil when value is nil.
func OrNil[T any](value *T) any {
	if value == nil {
		return nil
	}

	return *value
}
