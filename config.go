package microcache

import "time"

const cleanUpFactor = 0.01

type Configuration struct {
	MaxSize           int
	CleanupCount      int
	ExpireAfterWrite  time.Duration
	ExpireAfterAccess time.Duration
}

func DefaultConfiguration(maxSize int) Configuration {
	return Configuration{MaxSize: maxSize, CleanupCount: int(float64(maxSize) * cleanUpFactor)}
}
