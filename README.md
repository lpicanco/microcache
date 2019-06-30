# microcache
In memory cache library for golang

[![GoDoc](https://godoc.org/github.com/lpicanco/microcache?status.svg)](https://godoc.org/github.com/lpicanco/microcache)
[![Go Report Card](https://goreportcard.com/badge/github.com/lpicanco/microcache)](https://goreportcard.com/report/github.com/lpicanco/microcache)
[![GoCover](http://gocover.io/_badge/github.com/lpicanco/microcache)](http://gocover.io/github.com/lpicanco/microcache)


## How to use

Simple usage:

```go
import (
	"fmt"

	"github.com/lpicanco/microcache"
	"github.com/lpicanco/microcache/configuration"
)

func main() {
	cache := microcache.New(configuration.DefaultConfiguration(100))
	cache.Put(42, "answer")

	value, found := cache.Get(42)
	if found {
		fmt.Printf("Value: %v\n", value)
	}

	fmt.Printf("Cache len: %v\n", cache.Len())

	cache.Invalidate(42)
}
```

Custom options:

```go

import (
	"fmt"
	"time"

	"github.com/lpicanco/microcache"
	"github.com/lpicanco/microcache/configuration"
)

func main() {
	cache := microcache.New(configuration.Configuration{
		MaxSize:           10000,
		ExpireAfterWrite:  1 * time.Hour,
		ExpireAfterAccess: 10 * time.Minute,
		CleanupCount:      5,
	})
	cache.Put(42, "answer")

	value, found := cache.Get(42)
	if found {
		fmt.Printf("Value: %v\n", value)
	}

	fmt.Printf("Cache len: %v\n", cache.Len())

	cache.Invalidate(42)
}
```