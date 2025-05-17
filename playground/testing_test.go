package playground

import (
	"fmt"
	"testing"
	"time"

	clockwork "github.com/jonboulle/clockwork"
)

func TestMath(t *testing.T) {
	t.Skip()
	t.Run("1+2", func(t *testing.T) {
		val := 1+2
		if val != 3 {
			t.Errorf("Failed check that %s = %d", "1+2", 3)
		}

		t.Run("+2", func(t *testing.T) {
			val += 2
			if val != 5 {
				t.Errorf("Failed check that %s = %d", "+2", 5)
			}
			t.Run("+2", func(t *testing.T) {
				val += 2
				if val != 7 {
					t.Errorf("Failed check that %s = %d", "+2", 7)
				}
			})
		})

		t.Run("-2", func(t *testing.T) {
			val -= 2
			if val != 1 {
				t.Errorf("Failed check that %s = %d, it is %d", "-2", 1, val)
			}
		})
	})
	t.Run("2-1", func(t *testing.T) {
		val := 2-1
		if val != 1 {
			t.Errorf("Failed check that %s = %d", "2-1", 1)
		}
	})
}

func TestMyFunc(t *testing.T) {
	tickCount := 0
	c := clockwork.NewFakeClock()
	ticker := c.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Start goroutine to count ticks
	go func() {
		for range ticker.Chan() {
			fmt.Printf("incrementing %v\n", tickCount)
			tickCount++
		}
	}()

	// Advance the clock by 5.5 seconds while leaving time for tick counters
	step := 100*time.Millisecond
	for range (5500 * time.Millisecond/step) {
		c.Advance(step)
		time.Sleep(100 * time.Microsecond)
	}

	// Assert that tickCount is 5
	if tickCount != 5 {
		t.Errorf("Expected tickCount to be 5, but got %d", tickCount)
	}
}

func TestClockworkNow(t *testing.T) {
	c := clockwork.NewFakeClock()
	prevNow := c.Now()
	expected := prevNow.Add(1*time.Second).UnixNano()
	c.Advance(1*time.Second)
	now := c.Now().UnixNano()
	fmt.Printf("Expected %v and got %v which is definitely more than %v", expected, now, prevNow.UnixNano())
	if expected != now {
		t.Errorf("Expected time did not pass")
	}
}
