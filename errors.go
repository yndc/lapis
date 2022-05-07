package lapis

import "fmt"

// Indicates that the given key is not able to be resolved
type ErrNotFound[TKey any] struct {
	key TKey
}

func (m ErrNotFound[TKey]) Error() string {
	return fmt.Sprintf("not found: (%v)", m.key)
}

func NewErrNotFound[TKey any](key TKey) ErrNotFound[TKey] {
	return ErrNotFound[TKey]{
		key: key,
	}
}
