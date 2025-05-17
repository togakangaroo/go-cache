This is a simple, performant in-memory cache system with expiration functionality, implemented in Go.

# Features

- Thread-safe operations
- Key-value storage with automatic expiration
- Custom expiration times for individual items
- Ability to have an item never expire
- Automatic cleanup of expired items on a schedule
- Manually delete values
- Simple and clean API
- Developed using TDD with decent test coverage

# Not Features

Features that could be implemented eventually but are not handled here
- Special handling for complex objects that themselves might be executing code from other threads
- Advanced clean up strategies and eviction policies
- Refreshing a cache item based on its time of access
- Force cleanup
- Max capacitly limit and automated eviction of older items
- Sharding and distribution across multiple nodes
- Statistics tracking
- Extension points and hooks for events
- More expressive unit tests. Workflow tests like this should follow a tree structure, but the fact that go-test doesn't reset state between nested Run blocks state means implementing this is non-trivial and out of scope
- Load tests. While the thread safety mechanism here is quite simple, it would be a good idea to kick its tires by spinning up a test with loads of threads accessing the cache all at once. However, writing that test isn't completely straightforward and I'm happy to rely on the straightforwardness of the locking mechanism used at the moment

# Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/togakangaroo/go-cache-system/cache"
)
func main() {// Create a new cache with default expiration of 5 minutes and cleanup every 10 minutes
    c := cache.NewDefaultCache(5*time.Minute)
    defer c.Stop() // Don't forget to stop the cleanup goroutine

    // Set a value with the default expiration time
    c.Set("key1", "value1")

    // Set a value with a custom expiration time
    c.SetWithExpiration("key2", "value2", 1*time.Hour)

    // Set a value that never expires
    c.SetWithoutExpiration("key3", "value3")

    // Retrieve a value
    if val, found := c.Get("key1"); found {
        fmt.Printf("Found: %v\n", val)
    }

    // Delete a value
    c.Delete("key1")
}
```

# API Reference

## Creating a New Cache

```go
cache := cache.NewDefaultCache(defaultExpiration, cleanupInterval time.Duration)
```

- `defaultExpiration`: The default expiration time for items added to the cache. Use 0 for no expiration.
- `cleanupInterval`: The interval at which expired items are removed from the cache. Use 0 to disable automatic cleanup.

## Methods

- `Set(key string, value interface{})`: Add an item to the cache with the default expiration time.
- `SetWithExpiration(key string, value interface{}, expiration time.Duration)`: Add an item with a custom expiration time.
- `SetWithoutExpiration(key string, value interface{})`: Add an item with a custom expiration time.
- `Get(key string) (interface{}, bool)`: Retrieve an item from the cache. Returns the value and a boolean indicating if the key was found.
- `Delete(key string)`: Manually remove an item from the cache.
- `Stop()`: Stop the automatic cleanup goroutine. This should be run when you are cleaning up the Cache.

# Testing

The cache system includes comprehensive tests covering basic operations, custom expirations, thread safety, and edge cases.

Run the tests with:

```bash
go test ./cache
```

# Logging

There is some basic logging included. Run with the `LOG_LEVEL=DEBUG` environment variable set to see full logs
