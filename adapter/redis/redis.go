/*
MIT License

Copyright (c) 2018 Victor Springer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package redis

import (
	"context"
	"time"

	cache "github.com/rishikesh-parspec/echo-http-cache"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

// Adapter is the memory adapter data structure.
type Adapter struct {
	store *redisCache.Cache
}

// RingOptions exports go-redis RingOptions type.
type RingOptions redis.RingOptions

// Get implements the cache Adapter interface Get method.
func (a *Adapter) Get(key uint64) ([]byte, bool) {
	var c []byte
	if err := a.store.Get(context.Background(), cache.KeyAsString(key), &c); err == nil {
		return c, true
	}

	return nil, false
}

// Set implements the cache Adapter interface Set method.
func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) {
	a.store.Set(&redisCache.Item{
		Key:   cache.KeyAsString(key),
		Value: response,
		TTL:   expiration.Sub(time.Now()),
	})
}

// Release implements the cache Adapter interface Release method.
func (a *Adapter) Release(key uint64) {
	a.store.Delete(context.Background(), cache.KeyAsString(key))
}

// Purge implements the Adapter interface Purge method
func (a *Adapter) Purge() {
	panic("not implemented")
}

// NewAdapter initializes Redis adapter.
func NewAdapter(opt *RingOptions) cache.Adapter {
	ropt := redis.RingOptions(*opt)
	return &Adapter{
		redisCache.New(&redisCache.Options{
			Redis: redis.NewRing(&ropt),
		}),
	}
}
