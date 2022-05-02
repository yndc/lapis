package lapis

type LoadOptionType int

const (
	LoadOptionNoBatch = iota
)

// LoadOption is optional parameters to be added when loading a data, using a similar pattern with grpc.Options
type LoadOption interface {
	GetType() LoadOptionType
}

// NoBatch option will disable batching when loading a data by key
type NoBatch struct{}

// Check if NoBatch is present in the options
func UseNoBatch(options []LoadOption) bool {
	for _, v := range options {
		if v.GetType() == LoadOptionNoBatch {
			return true
		}
	}
	return false
}

func (o NoBatch) GetType() LoadOptionType { return LoadOptionNoBatch }
