package lapis

import "context"

// Load a data from it's key
func (r *Store[TKey, TValue]) Load(key TKey, flags ...LoadFlag) (TValue, error) {
	if !r.useBatcher || hasLoadFlag(r.defaultLoadFlags, flags, LoadNoBatch) {
		return singlify(r.resolveAndCollect)(key)
	}
	return r.batcher.Load(key)
}

// Load a data by key with a context, if the context is cancelled, the data loading will be cancelled too if
// this is the only operation
func (r *Store[TKey, TValue]) LoadCtx(ctx context.Context, key TKey, flags ...[]int) (TValue, error) {
	panic("not implemented")
}

// Load a set of data from their keys
func (r *Store[TKey, TValue]) LoadAll(keys []TKey, flags ...LoadFlag) ([]TValue, []error) {
	if !r.useBatcher || hasLoadFlag(r.defaultLoadFlags, flags, LoadNoBatch) {
		return r.resolveAndCollect(keys)
	}
	return r.batcher.LoadAll(keys)
}
