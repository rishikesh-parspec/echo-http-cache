# echo-http-cache
[![Build Status](https://travis-ci.org/victorspringer/http-cache.svg?branch=master)](https://travis-ci.org/victorspringer/http-cache) [![Coverage Status](https://coveralls.io/repos/github/victorspringer/http-cache/badge.svg?branch=master)](https://coveralls.io/github/victorspringer/http-cache?branch=master) [![](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat)](https://godoc.org/github.com/SporkHubr/echo-http-cache)

This is a high performance Golang HTTP middleware for server-side application layer caching, ideal for REST APIs, using Echo framework.

It is simple, super fast, thread safe and gives the possibility to choose the adapter (memory, Redis, DynamoDB etc).

The memory adapter minimizes GC overhead to near zero and supports some options of caching algorithms (LRU, MRU, LFU, MFU). This way, it is able to store plenty of gigabytes of responses, keeping great performance and being free of leaks.

## Updating the package
1. Make your changes and commit.
2. `git tag vX.Y.Z`
3. `git push origin vX.Y.Z`
4. `GOPROXY=proxy.golang.org go list -m github.com/parikshit-parspec/echo-http-cache@vX.Y.Z`
5. In the dependent project update version in `go.mod` and run `go mod tidy`.
   
## Getting Started

### Installation (Go Modules)
`go get github.com/SporkHubr/echo-http-cache`

### Usage
This is an example of use with the memory adapter:

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"
    
    "github.com/SporkHubr/echo-http-cache"
    "github.com/SporkHubr/echo-http-cache/adapter/memory"
    "github.com/labstack/echo/v4"
)

func example(c echo.Context) {
   c.String(http.StatusOk, "Ok")
}

func main() {
    memcached, err := memory.NewAdapter(
        memory.AdapterWithAlgorithm(memory.LRU),
        memory.AdapterWithCapacity(10000000),
    )
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    cacheClient, err := cache.NewClient(
        cache.ClientWithAdapter(memcached),
        cache.ClientWithTTL(10 * time.Minute),
        cache.ClientWithRefreshKey("opn"),
    )
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    router := echo.New()
    router.Use(cacheClient.Middleware())
    router.GET("/", example)
    e.Start(":8080")
}
```

Example of Client initialization with Redis adapter:
```go
import (
    "github.com/SporkHubr/echo-http-cache"
    "github.com/SporkHubr/echo-http-cache/adapter/redis"
)

...

    ringOpt := &redis.RingOptions{
        Addrs: map[string]string{
            "server": ":6379",
        },
    }
    cacheClient := cache.NewClient(
        cache.ClientWithAdapter(redis.NewAdapter(ringOpt)),
        cache.ClientWithTTL(10 * time.Minute),
        cache.ClientWithRefreshKey("opn"),
    )

...
```

## Benchmarks
The benchmarks were based on [allegro/bigache](https://github.com/allegro/bigcache) tests and used to compare it with the http-cache memory adapter.<br>
The tests were run using an Intel i5-2410M with 8GB RAM on Arch Linux 64bits.<br>
The results are shown below:

### Writes and Reads
```bash
cd adapter/memory/benchmark
go test -bench=. -benchtime=10s ./... -timeout 30m

BenchmarkHTTPCacheMamoryAdapterSet-4             5000000     343 ns/op    172 B/op    1 allocs/op
BenchmarkBigCacheSet-4                           3000000     507 ns/op    535 B/op    1 allocs/op
BenchmarkHTTPCacheMamoryAdapterGet-4            20000000     146 ns/op      0 B/op    0 allocs/op
BenchmarkBigCacheGet-4                           3000000     343 ns/op    120 B/op    3 allocs/op
BenchmarkHTTPCacheMamoryAdapterSetParallel-4    10000000     223 ns/op    172 B/op    1 allocs/op
BenchmarkBigCacheSetParallel-4                  10000000     291 ns/op    661 B/op    1 allocs/op
BenchmarkHTTPCacheMemoryAdapterGetParallel-4    50000000    56.1 ns/op      0 B/op    0 allocs/op
BenchmarkBigCacheGetParallel-4                  10000000     163 ns/op    120 B/op    3 allocs/op
```
http-cache writes are slightly faster and reads are much more faster.

### Garbage Collection Pause Time
```bash
cache=http-cache go run benchmark_gc_overhead.go

Number of entries:  20000000
GC pause for http-cache memory adapter:  2.445617ms

cache=bigcache go run benchmark_gc_overhead.go

Number of entries:  20000000
GC pause for bigcache:  7.43339ms
```
echo-http-cache memory adapter takes way less GC pause time, that means smaller GC overhead.

## Roadmap
- Make it compliant with RFC7234
- Add more middleware configuration (cacheable status codes, paths etc)
- Develop gRPC middleware
- Develop Badger adapter
- Develop DynamoDB adapter
- Develop MongoDB adapter

## Godoc Reference
- [echo-http-cache](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache)
- [Memory adapter](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache/adapter/memory)
- [Redis adapter](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache/adapter/redis)

## License
echo-http-cache is released under the [MIT License](https://github.com/SporkHubr/echo-http-cache/blob/master/LICENSE).
