package main

import (
	"sync"
	"testing"
	"time"
)

func TestBroadcaster(t *testing.T) {
	subscriberCount := 5
	msg := 1

	b := NewBroadcaster[int]()
	go b.Start()
	defer b.Stop()

    // This will hold the messages that each subscriber received, indexed by
    // subscriber number.
	results := map[int]int{}
	resultsLock := sync.Mutex{}

    // Create a simple subscriber function
	sub := func(i int) {
		c := b.Subscribe()
		result := <-c

        // Handle concurrent writes to the results map
		resultsLock.Lock()
		results[i] = result
		resultsLock.Unlock()

		if result != msg {
			t.Errorf("Expected to receive %v from channel %v, got %v", msg, i, result)
		}
	}

	// Start multiple subscribers
	for i := range subscriberCount {
		go sub(i)
	}

	// Give time for subscribers to start
	time.Sleep(50 * time.Millisecond)

	b.Publish(msg)

	// Give time for subscribers to receive message
	time.Sleep(50 * time.Millisecond)

	// Assert that all subscribers received the message
	for i := range subscriberCount {
		r := results[i]
		if r != msg {
			t.Errorf("Expected to receive %v from channel %v, but got %v", msg, i, r)
		}
	}
}
