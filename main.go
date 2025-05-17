package main

import (
	"fmt"
	"time"
	"github.com/togakangaroo/go-cache-system/cache"
)
func main() {
	// Create a new cache with default expiration of 5 minutes and cleanup every 10 minutes
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
