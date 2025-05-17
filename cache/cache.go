package cache

import (
	"log/slog"
	"os"
	"sync"
	"time"

	clockwork "github.com/jonboulle/clockwork"
)

type Cache struct {
	items             map[string]Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan any
	clock             clockwork.Clock
	mu                sync.RWMutex
	logger            *slog.Logger
}

// Use NewDefaultCache or this to create a Cache instance, prefer not to refernece it directly
func NewCache(defaultExpiration, cleanupInterval time.Duration, clock clockwork.Clock) *Cache {
	cache := &Cache{
		items:             make(map[string]Item),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan any),
		clock:             clock,
		logger:            setupLogger(),
	}

	if cleanupInterval > 0 {
		go cache.startCleanupTimer()
	}

	return cache
}
// Create a cache with some obvious defaults set. This is probably the version that you want.
// See NewCache for more complex version
func NewDefaultCache(defaultExpiration time.Duration) *Cache {
	return NewCache(defaultExpiration, 30 * time.Second, clockwork.NewRealClock())
}

// Add an item to the cache with the default expiration time.
// If the item already exists in the cache it will be overridden and its expiration time updated.
func (c *Cache) Set(key string, value any) {
	c.SetWithExpiration(key, value, c.defaultExpiration)
}

// Add an item to the cache with a custom expiration time.
// If expiration is 0, the item never expires.
func (c *Cache) SetWithExpiration(key string, value any, expiration time.Duration) {
	var exp int64
	if 0 < expiration {
		exp = c.clock.Now().Add(expiration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item{
		Value:      value,
		Expiration: exp,
	}
}

// Adds an item to the cache that never expires
func (c *Cache) SetWithoutExpiration(key string, value any) {
	c.SetWithExpiration(key, value, 0)
}

// Manually remove an item from the cache. You will usually not have to do this and can either overwrite values in the cache and/or wait for them to expire
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Return cached item and a boolean indicating whether the key was found
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if 0 < item.Expiration && item.Expiration < c.clock.Now().UnixNano() {
		c.logger.Debug("Item with key `", key, "` found but expired. Not returning.")
		return nil, false
	}

	return item.Value, true
}

type Item struct {
	Value      any
	Expiration int64
}

func setupLogger() *slog.Logger {
	logger := slog.Default()
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		var logLevel slog.Level
		switch logLevelStr {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		default:
			// Default to INFO if unrecognized level
			logLevel = slog.LevelInfo
		}

		loggerOpts := &slog.HandlerOptions{Level: logLevel}
		logger = slog.New(slog.NewTextHandler(os.Stderr, loggerOpts))
	}
	return logger
}

// starts the cleanup timer
func (c *Cache) startCleanupTimer() {
	ticker := c.clock.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.Chan():
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// Force eviction of all expired items
func (c *Cache) deleteExpired() {
	c.logger.Debug("Cleaning up expired items.")
	now := c.clock.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if 0 < v.Expiration && v.Expiration < now {
			delete(c.items, k)
		}
	}
}

// Stops the cleanup and properly disposes of the cache
func (c *Cache) Stop() {
	if 0 < c.cleanupInterval {
		c.stopCleanup <- true
	}
}
