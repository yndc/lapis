package lapis

import (
	"time"
)

// a batch of keys to be resolved
type batch[TKey comparable, TValue any] struct {
	keys    []TKey
	data    []TValue
	errors  []error
	done    []chan struct{} // channels to notify if any of each keys in this batch has been resolved
	allDone chan struct{}   // channel to notify that all keys in this batch has been resolved
	closing bool            // flags this batch as closed, no more keys can be added in this batch
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch, if it is the first key in the batch, then the
// batch timer will be started
func (b *batch[TKey, TValue]) keyIndex(l *Batcher[TKey, TValue], key TKey) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	b.done = append(b.done, make(chan struct{}))
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.pendingBatch = nil
			go b.resolveBatch(l)
		}
	}

	return pos
}

func (b *batch[TKey, TValue]) startTimer(l *Batcher[TKey, TValue]) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.pendingBatch = nil
	l.mu.Unlock()

	b.resolveBatch(l)
}

func (b *batch[TKey, TValue]) resolveBatch(l *Batcher[TKey, TValue]) {
	b.data = make([]TValue, len(b.keys))
	b.errors = make([]error, len(b.keys))
	l.resolver(b.keys, b.finishKey)
	close(b.allDone)
	l.mu.Lock()
	for _, key := range b.keys {
		delete(l.batches, key)
	}
	l.mu.Unlock()
}

func (b *batch[TKey, TValue]) finishKey(index int, value TValue, err error) {
	b.data[index] = value
	b.errors[index] = err
	close(b.done[index])
}
