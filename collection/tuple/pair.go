package tuple

type Pair[T1 any, T2 any] struct {
	V1 T1
	V2 T2
}

func NewPair[T1 any, T2 any](v1 T1, v2 T2) Pair[T1, T2] {
	return Pair[T1, T2]{
		V1: v1,
		V2: v2,
	}
}
