package cache

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/blend/go-sdk/async"
)

var (
	_ Cache = (*LocalCache)(nil)
)

// NewLocalCache returns a new LocalLocalCache.
func NewLocalCache(options ...LocalCacheOption) *LocalCache {
	c := LocalCache{
		Data: make(map[interface{}]*Value),
	}
	for _, opt := range options {
		opt(&c)
	}
	return &c
}

// LocalCacheOption is a local cache option.
type LocalCacheOption func(*LocalCache)

// OptLocalCacheSweepInterval sets the local cache sweep interval.
func OptLocalCacheSweepInterval(d time.Duration) LocalCacheOption {
	return func(lc *LocalCache) {
		lc.Sweeper = async.NewInterval(lc.Sweep, d)
	}
}

// LocalCache is a memory LocalCache.
type LocalCache struct {
	sync.RWMutex
	Data    map[interface{}]*Value
	Sweeper *async.Interval
}

// Start starts the sweeper.
func (lc *LocalCache) Start() error {
	if lc.Sweeper == nil {
		return nil
	}
	return lc.Sweeper.Start()
}

// NotifyStarted returns the underlying started signal.
func (lc *LocalCache) NotifyStarted() <-chan struct{} {
	if lc.Sweeper == nil {
		return nil
	}
	return lc.Sweeper.NotifyStarted()
}

// Stop stops the sweeper.
func (lc *LocalCache) Stop() error {
	if lc.Sweeper == nil {
		return nil
	}
	return lc.Sweeper.Stop()
}

// NotifyStopped returns the underlying stopped signal.
func (lc *LocalCache) NotifyStopped() <-chan struct{} {
	if lc.Sweeper == nil {
		return nil
	}
	return lc.Sweeper.NotifyStopped()
}

// Sweep checks keys for expired ttls.
// If any values are configured with 'OnSweep' handlers, they will be called
// outside holding the critical section.
func (lc *LocalCache) Sweep(ctx context.Context) error {
	lc.Lock()
	now := time.Now().UTC()
	var keysToRemove []interface{}
	var handlers []func()
	for key, value := range lc.Data {
		if now.After(value.Timestamp.Add(value.TTL)) {
			keysToRemove = append(keysToRemove, key)
			if value.OnSweep != nil {
				handlers = append(handlers, value.OnSweep)
			}
		}
	}
	for _, key := range keysToRemove {
		delete(lc.Data, key)
	}
	lc.Unlock()

	// call the handlers outside the critical section.
	for _, handler := range handlers {
		handler()
	}
	return nil
}

// Set adds a LocalCache item.
func (lc *LocalCache) Set(key, value interface{}, options ...ValueOption) {
	if key == nil {
		panic("nil key")
	}

	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	if reflect.TypeOf(key).Kind() != reflect.Struct {
		panic("key is not a struct; consider using a key type to leverage the compiler for key checking")
	}

	v := Value{
		Timestamp: time.Now().UTC(),
		Key:       key,
		Value:     value,
	}

	for _, opt := range options {
		opt(&v)
	}

	lc.Lock()
	if lc.Data == nil {
		lc.Data = make(map[interface{}]*Value)
	}
	lc.Data[key] = &v
	lc.Unlock()
}

// Get gets a value based on a key.
func (lc *LocalCache) Get(key interface{}) (value interface{}, hit bool) {
	lc.RLock()
	valueNode, ok := lc.Data[key]
	lc.RUnlock()

	if ok {
		value = valueNode.Value
		hit = true
		return
	}
	return
}

// Has returns if the key is present in the LocalCache.
func (lc *LocalCache) Has(key interface{}) (has bool) {
	lc.RLock()
	_, has = lc.Data[key]
	lc.RUnlock()
	return
}

// Remove removes a specific key.
func (lc *LocalCache) Remove(key interface{}) (value interface{}, hit bool) {
	lc.Lock()
	valueData, ok := lc.Data[key]
	lc.Unlock()

	if !ok {
		return
	}

	value = valueData.Value
	hit = true

	if valueData.OnRemove != nil {
		valueData.OnRemove()
	}
	delete(lc.Data, key)
	return
}

// Stats returns the LocalCache stats.
func (lc *LocalCache) Stats() (stats Stats) {
	lc.RLock()
	defer lc.RUnlock()

	stats.Count = len(lc.Data)
	now := time.Now().UTC()
	for _, item := range lc.Data {
		age := now.Sub(item.Timestamp)
		if stats.MaxAge < age {
			stats.MaxAge = age
		}
		stats.SizeBytes += int(reflect.TypeOf(item).Size())
	}
	return
}
