package layer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/flowscan/lapis"
	"github.com/mediocregopher/radix/v3"
	"github.com/rs/zerolog/log"
)

// A hard-coded constant for nil values to differentiate nil and undefined (not found) values
const RedisNilValue = "__@@@__LAPIS_REDIS_NIL_VALUE"

// Configuration for the redis data layer
type RedisConfig struct {
	// The duration of the cached data, set 0 to disable expiration
	Retention time.Duration

	// Connection to redis
	Connection *radix.Pool

	// Key prefix to be used in redis keys
	KeyPrefix string
}

// RedisGob layer is redis-backed cache layer with gob encoding and configurable expiration time
type RedisGob[TKey comparable, TValue any] struct {
	config RedisConfig
}

// Unique identifier for this layer used for logging and metric purposes
func (l *RedisGob[TKey, TValue]) Identifier() string { return "redis" }

// The function that will be used to resolve a set of keys
func (l *RedisGob[TKey, TValue]) Get(keys []TKey) ([]TValue, []error) {
	keysCount := len(keys)
	result := make([]TValue, keysCount)
	errors := make([]error, keysCount)
	cacheBuffer := make([][]byte, keysCount)
	if err := l.config.Connection.Do(radix.Cmd(&cacheBuffer, "MGET", stringifyKeys(keys, l.config.KeyPrefix)...)); err != nil {
		fillArray(errors, err)
	} else {
		for i, k := range keys {
			if cacheBuffer[i] != nil {
				// handle nil values since it is defined with a constant
				if string(cacheBuffer[i]) == RedisNilValue {
					var zeroValue TValue
					result[i] = zeroValue
				}
				buffer := bytes.NewBuffer(cacheBuffer[i])
				if err := gob.NewDecoder(buffer).Decode(&result[i]); err != nil {
					errors[i] = err
				}
			} else {
				errors[i] = lapis.NewErrNotFound(k)
			}
		}
	}

	return result, errors
}

// The function that will be called for successful resolvers
func (l *RedisGob[TKey, TValue]) Set(keys []TKey, values []TValue) []error {
	count := len(keys)

	// prepare batch SET commands using MSET
	cacheArguments := make([]string, 2*count)
	keysString := stringifyKeys(keys, l.config.KeyPrefix)
	for i, value := range values {
		// handle nil values
		// gob does not support nil pointers unfortunately
		if reflect.TypeOf(value).Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil() {
			// we set the hard-coded nil constant for nil values
			cacheArguments[2*i+1] = RedisNilValue
		} else {
			b := bytes.Buffer{}
			err := gob.NewEncoder(&b).Encode(value)
			if err != nil {
				log.Err(err).Send()
				continue
			}
			cacheArguments[2*i+1] = b.String()
		}
		cacheArguments[2*i] = keysString[i]
	}
	commands := make([]radix.CmdAction, 1+count)
	commands[0] = radix.Cmd(nil, "MSET", cacheArguments...)

	// prepare EXPIRE commands
	if l.config.Retention > 0 {
		for i, key := range keysString {
			commands[i+1] = radix.Cmd(nil, "EXPIRE", key, strconv.FormatInt(int64(l.config.Retention.Seconds()), 10))
		}
	}
	err := l.config.Connection.Do(radix.Pipeline(commands...))
	if err != nil {
		log.Err(err).Send()
	}
	return nil
}

// Create a new redis data layer
func NewRedis[TKey comparable, TValue any](config RedisConfig) *RedisGob[TKey, TValue] {
	l := &RedisGob[TKey, TValue]{
		config: config,
	}
	return l
}

func stringifyKeys[TKey comparable](keys []TKey, prefix string) []string {
	return mapFn(keys, func(input TKey) string {
		return fmt.Sprintf("%s%v", prefix, input)
	})
}

func mapFn[T1 any, T2 any](arr []T1, fn func(input T1) T2) []T2 {
	newArr := make([]T2, len(arr))
	for i, v := range arr {
		newArr[i] = fn(v)
	}
	return newArr
}

func fillArray[T any](arr []T, value T) []T {
	for i := range arr {
		arr[i] = value
	}
	return arr
}
