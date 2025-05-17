package cache

import (
	"fmt"
	"testing"
	"time"

	clockwork "github.com/jonboulle/clockwork"
)

type CacheTestContext struct {
	Cache *Cache
	Clock *clockwork.FakeClock
	t *testing.T
}

func (ctx CacheTestContext) WithT(t *testing.T) CacheTestContext {
	return CacheTestContext{
		Cache: ctx.Cache,
		Clock: ctx.Clock,
		t:     t,
	}
}

func assertKeyAndExpectedValueInCache(key string, expected any, x CacheTestContext) {
	val, found := x.Cache.Get(key)

	x.t.Run(fmt.Sprintf("THEN `%s` is found in the cache", key), func(t *testing.T) {
		if !found {
			t.Errorf("Expected to find cached value under key `%s` but one was not found", key)
		}
	})
	x.t.Run(fmt.Sprintf("THEN value at `%s` is expected", key), func(t *testing.T) {
		if val != expected {
			t.Errorf("Value is different than expected, [`%s`] = `%v`", key, val)
		}
	})
}

func assertKeyNotInCache(key string, x CacheTestContext) {
	x.t.Run(fmt.Sprintf("THEN `%s` is not found in the cache", key), func(t *testing.T) {
		if _, found := x.Cache.Get(key); found {
			t.Errorf("Expected to not find a cached value under key `%s` but one was found", key)
		}
	})
}
func elapses(duration time.Duration, x CacheTestContext, fn func(x CacheTestContext)) {
	x.t.Run(fmt.Sprintf("when %v elapses", duration), func(t *testing.T) {
		x.Clock.Advance(duration)
		fn(x.WithT(t))
	})
}

func TestCacheBasicOperations(t *testing.T) {
	withExpiringCacheOf1s := func (fn func(x CacheTestContext)) {
		t.Run("When we have a cache with default expiration of `1s`", func(t *testing.T){
			clock := clockwork.NewFakeClock()
			cache := NewCache(1*time.Second, 2*time.Second, clock)
			defer cache.Stop()
			fn(CacheTestContext{
				Cache: cache,
				Clock: clock,
				t:     t,
			})
		})
	}

	// Getting items
	withExpiringCacheOf1s(func (x CacheTestContext) {
		x.t.Run("when user puts an items under the keys `foo` and `bar` into the cache", func(t *testing.T){
			x.Cache.Set("foo", "foo value")
			x.Cache.Set("bar", "bar value")
			assertKeyAndExpectedValueInCache("foo", "foo value", x.WithT(t))
			assertKeyAndExpectedValueInCache("bar", "bar value", x.WithT(t))
		})
	})

	// Deleting items
	withExpiringCacheOf1s(func (x CacheTestContext) {
		x.t.Run("when user puts an item under the key `foo` into the cache", func(t *testing.T){
			x.Cache.Set("foo", "foo value")
			t.Run("when user deletes the key `foo`", func(t *testing.T){
				x.Cache.Delete("foo")
				assertKeyNotInCache("foo", x.WithT(t))
			})
		})
	})

	// Item expiration
	withExpiringCacheOf1s(func (x CacheTestContext) {
		t.Run("when user puts an item `foo` into the cache", func(t *testing.T){
			x.Cache.Set("foo", "foo value")
			elapses(900 * time.Millisecond, x, func(x CacheTestContext) {
				assertKeyAndExpectedValueInCache("foo", "foo value", x)
				elapses(200 * time.Millisecond, x, func(x CacheTestContext) {
					assertKeyNotInCache("foo", x)
				})
			})
		})
	})
	withExpiringCacheOf1s(func (x CacheTestContext) {
		t.Run("when user puts an item `foo` into the cache", func(t *testing.T){
			x.Cache.Set("foo", "foo value")
			elapses(900 * time.Millisecond, x, func(x CacheTestContext) {
				t.Run("when user resets the expiration by putting an item `foo` into the cache again", func(t *testing.T){
					x.Cache.Set("foo", "foo value 2")
					assertKeyAndExpectedValueInCache("foo", "foo value 2", x)
					elapses(200 * time.Millisecond, x, func(x CacheTestContext) {
						assertKeyAndExpectedValueInCache("foo", "foo value 2", x)
					})
				})
			})
		})
	})
	withExpiringCacheOf1s(func (x CacheTestContext) {
		t.Run("when user puts an item `foo` into the cache with a custom expiration of 2s", func(t *testing.T){
			x.Cache.SetWithExpiration("foo", "foo value", 2 * time.Second)
			elapses(1900 * time.Millisecond, x, func(x CacheTestContext) {
				assertKeyAndExpectedValueInCache("foo", "foo value", x)
				elapses(200 * time.Millisecond, x, func(x CacheTestContext) {
					assertKeyNotInCache("foo", x)
				})
			})
		})
	})
	withExpiringCacheOf1s(func (x CacheTestContext) {
		t.Run("when user puts an item `foo` into the cache without expiration", func(t *testing.T){
			x.Cache.SetWithoutExpiration("foo", "foo value")
			elapses(10 * time.Second, x, func(x CacheTestContext) {
				assertKeyAndExpectedValueInCache("foo", "foo value", x)
			})
		})
	})
}

func TestCacheCleanup(t *testing.T) {
	t.Run("When we have a cache with default expiration of 1s and cleanup interval of `2s`", func(t *testing.T) {
		clock := clockwork.NewFakeClock()
		cache := NewCache(1*time.Second, 2*time.Second, clock)
		defer cache.Stop()

		// Add items to the cache
		cache.Set("item1", "value1")
		cache.Set("item2", "value2")

		cache.mu.RLock()
		if len(cache.items) != 2 {
			t.Errorf("Expected 2 items in cache, got %d", len(cache.items))
		}
		cache.mu.RUnlock()

		// Advance time past expiration but before cleanup
		step := 100*time.Millisecond
		for range (1100 * time.Millisecond/step) {
			clock.Advance(step)
			time.Sleep(100 * time.Microsecond)
		}

		val, found := cache.Get("item1")
		if found {
			t.Errorf("Expected item1 to be expired, but got %v", val)
		}

		cache.mu.RLock()
		if len(cache.items) != 2 {
			t.Errorf("Expected 2 items still in internal map before cleanup, got %d", len(cache.items))
		}
		cache.mu.RUnlock()

		// Advance time to trigger cleanup
		for range (1000 * time.Millisecond/step) {
			clock.Advance(step)
			time.Sleep(100 * time.Microsecond)
		}

		cache.mu.RLock()
		if len(cache.items) != 0 {
			t.Errorf("Expected 0 items in cache after cleanup, got %d", len(cache.items))
		}
		cache.mu.RUnlock()
	})
}
