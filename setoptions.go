package lapis

import "github.com/flowscan/lapis/ptr"

type SetOptionType int

const (
	SetOptionSequential = iota
	SetOptionAscending
)

// SetOption is optional parameters to be added when loading a data, using a similar pattern with grpc.Options
type SetOption interface {
	GetType() SetOptionType
}

// Run the set operation sequentially.
// The set operation for the next layer won't continue before the current layer's set procedure has been finished.
// By default the set operation will start from the last layer to the first layer.
type Sequential struct{}

func (o Sequential) GetType() SetOptionType { return SetOptionSequential }

// Run the set operation in ascending order, from the first layer to the last.
type Ascending struct{}

func (o Ascending) GetType() SetOptionType { return SetOptionAscending }

func GetOption[T any](source []SetOption, optionType int) *T {
	for _, v := range source {
		if int(v.GetType()) == optionType {
			return ptr.To(v.(T))
		}
	}
	return nil
}
