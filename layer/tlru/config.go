package tlru

import "time"

// configuration for a TLRU cache
type Config struct {
	MaxBytes   int64         // Maximum bytes in the storage, set to -1 to disable byte limit
	MaxItems   int           // Maximum number of items in the storage, set to -1 to disable item limit
	DefaultTTL time.Duration // Default TTL for items added into the cache, note that this value is only added as a weight into the LRU algorithm
}

const ConfigDefaultMaxBytes = 67_108_864
const ConfigDefaultMaxItems = 65_536

func (c *Config) Validate() error {
	if c.MaxBytes == 0 {
		c.MaxBytes = ConfigDefaultMaxBytes
	}
	if c.MaxItems == 0 {
		c.MaxItems = ConfigDefaultMaxItems
	}
	return nil
}
