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

package memory

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	cache "github.com/rishikesh-parspec/echo-http-cache"
	"sync"
	"time"
)

// Algorithm is the string type for caching algorithms labels.
type Algorithm string

const (
	// LRU is the constant for Least Recently Used.
	LRU Algorithm = "LRU"

	// MRU is the constant for Most Recently Used.
	MRU Algorithm = "MRU"

	// LFU is the constant for Least Frequently Used.
	LFU Algorithm = "LFU"

	// MFU is the constant for Most Frequently Used.
	MFU Algorithm = "MFU"
)

type Response struct {
	// Value is the cached response value.
	Value []byte

	// Expiration is the cached response expiration date.
	Expiration time.Time

	// LastAccess is the last date a cached response was accessed.
	// Used by LRU and MRU algorithms.
	LastAccess time.Time

	// Frequency is the count of times a cached response is accessed.
	// Used for LFU and MFU algorithms.
	Frequency int
}

// Adapter is the memory adapter data structure.
type Adapter struct {
	mutex     sync.RWMutex
	capacity  int
	algorithm Algorithm
	store     map[uint64][]byte
}

// AdapterOptions is used to set Adapter settings.
type AdapterOptions func(a *Adapter) error

// BytesToResponse converts bytes array into Response data structure.
func BytesToResponse(b []byte) Response {
	var r Response
	dec := gob.NewDecoder(bytes.NewReader(b))
	dec.Decode(&r)

	return r
}

// Bytes converts Response data structure into bytes array.
func (r Response) Bytes() []byte {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	enc.Encode(&r)

	return b.Bytes()
}

// Get implements the cache Adapter interface Get method.
func (a *Adapter) Get(key uint64) ([]byte, bool) {
	a.mutex.RLock()
	res, ok := a.store[key]
	a.mutex.RUnlock()
	if !ok {
		return nil, false
	}

	response := BytesToResponse(res)
	if response.Expiration.After(time.Now()) { // Cache is still valid
		response.LastAccess = time.Now()
		response.Frequency++
		a.mutex.Lock()
		a.store[key] = response.Bytes()
		a.mutex.Unlock()
		return response.Value, true
	}

	// Cache is expired, remove it
	a.Release(key)
	return nil, false
}

// Set implements the cache Adapter interface Set method.
func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) {
	now := time.Now()
	res := Response{
		Value:      response,
		Expiration: expiration,
		LastAccess: now,
		Frequency:  1,
	}
	a.mutex.RLock()
	length := len(a.store)
	a.mutex.RUnlock()
	if length > 0 && length == a.capacity {
		a.evict()
	}

	a.mutex.Lock()
	a.store[key] = res.Bytes()
	a.mutex.Unlock()
}

// Release implements the Adapter interface Release method.
func (a *Adapter) Release(key uint64) {
	a.mutex.RLock()
	_, ok := a.store[key]
	a.mutex.RUnlock()

	if ok {
		a.mutex.Lock()
		delete(a.store, key)
		a.mutex.Unlock()
	}
}

// Purge implements the Adapter interface Purge method
func (a *Adapter) Purge() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.store = make(map[uint64][]byte)
}

func (a *Adapter) evict() {
	selectedKey := uint64(0)
	lastAccess := time.Now()
	frequency := 2147483647

	if a.algorithm == MRU {
		lastAccess = time.Time{}
	} else if a.algorithm == MFU {
		frequency = 0
	}

	for k, v := range a.store {
		r := cache.BytesToResponse(v)
		switch a.algorithm {
		case LRU:
			if r.LastAccess.Before(lastAccess) {
				selectedKey = k
				lastAccess = r.LastAccess
			}
		case MRU:
			if r.LastAccess.After(lastAccess) ||
				r.LastAccess.Equal(lastAccess) {
				selectedKey = k
				lastAccess = r.LastAccess
			}
		case LFU:
			if r.Frequency < frequency {
				selectedKey = k
				frequency = r.Frequency
			}
		case MFU:
			if r.Frequency >= frequency {
				selectedKey = k
				frequency = r.Frequency
			}
		}
	}

	a.Release(selectedKey)
}

// NewAdapter initializes memory adapter.
func NewAdapter(opts ...AdapterOptions) (cache.Adapter, error) {
	a := &Adapter{}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	if a.capacity <= 1 {
		return nil, errors.New("memory adapter capacity is not set")
	}

	if a.algorithm == "" {
		return nil, errors.New("memory adapter caching algorithm is not set")
	}

	a.mutex = sync.RWMutex{}
	a.store = make(map[uint64][]byte, a.capacity)

	return a, nil
}

// AdapterWithAlgorithm sets the approach used to select a cached
// response to be evicted when the capacity is reached.
func AdapterWithAlgorithm(alg Algorithm) AdapterOptions {
	return func(a *Adapter) error {
		a.algorithm = alg
		return nil
	}
}

// AdapterWithCapacity sets the maximum number of cached responses.
func AdapterWithCapacity(cap int) AdapterOptions {
	return func(a *Adapter) error {
		if cap <= 1 {
			return fmt.Errorf("memory adapter requires a capacity greater than %v", cap)
		}

		a.capacity = cap

		return nil
	}
}
