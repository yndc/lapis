package lapis

type SetFlag int

const (
	SetSequential SetFlag = 1 << iota // Run the set operation sequentially
	SetAscending                      // Run the set operation in ascending order, from the first layer to the last.
)

// check if the given flags is enabled
func hasSetFlag(def SetFlag, flags []SetFlag, flag SetFlag) bool {
	sum := def
	for _, f := range flags {
		sum = sum | f
	}
	return (sum & flag) == flag
}
