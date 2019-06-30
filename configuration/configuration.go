package configuration

import "time"

const cleanUpFactor = 0.01

// Configuration for microcache
type Configuration struct {
	// Max size of the cache
	MaxSize int
	// Number of items to cleanup when the cache becomes full
	CleanupCount int

	// Time to expire items after creation
	ExpireAfterWrite time.Duration

	// Time to expire items after last access
	ExpireAfterAccess time.Duration
}

// DefaultConfiguration with maxSize setting
func DefaultConfiguration(maxSize int) Configuration {
	return Configuration{MaxSize: maxSize, CleanupCount: int(float64(maxSize) * cleanUpFactor)}
}
