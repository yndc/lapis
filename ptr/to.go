package ptr

func To[T any](value T) *T {
	cp := value
	return &cp
}
