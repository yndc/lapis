package lapis

type LoadFlag int

const (
	LoadNoBatch        LoadFlag = 1 << iota // Don't use the batcher when loading
	LoadNoCollectBatch                      // Don't use collector on the batcher
	LoadNoShareBatch                        // Don't use an existing ongoing batch
)

// check if the given flags is enabled
func hasLoadFlag(def LoadFlag, flags []LoadFlag, flag LoadFlag) bool {
	sum := def
	for _, f := range flags {
		sum = sum | f
	}
	return (sum & flag) == flag
}
