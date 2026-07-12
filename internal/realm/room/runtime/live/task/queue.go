// Package task schedules cancelable work owned by one active room.
package task

import (
	"sync"
	"time"
)

// Key identifies replaceable scheduled work.
type Key uint64

// Run executes scheduled room work.
type Run func(time.Time)

// entry stores one pending task.
type entry struct {
	// key identifies replaceable work and remains zero for independent tasks.
	key Key
	// deadline stores the earliest execution time.
	deadline time.Time
	// run executes the task.
	run Run
}

// Queue stores pending room-owned tasks.
type Queue struct {
	// mutex protects pending tasks.
	mutex sync.Mutex
	// entries stores pending tasks in insertion order.
	entries []entry
}

// Schedule appends independent work.
func (queue *Queue) Schedule(deadline time.Time, run Run) {
	queue.schedule(0, deadline, run, false)
}

// Replace schedules work after removing an existing task with the same key.
func (queue *Queue) Replace(key Key, deadline time.Time, run Run) {
	queue.schedule(key, deadline, run, true)
}

// Due removes and returns work whose deadline has passed.
func (queue *Queue) Due(now time.Time) []Run {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if len(queue.entries) == 0 {
		return nil
	}
	due := make([]Run, 0, len(queue.entries))
	pending := queue.entries[:0]
	for _, scheduled := range queue.entries {
		if scheduled.deadline.After(now) {
			pending = append(pending, scheduled)
			continue
		}
		due = append(due, scheduled.run)
	}
	queue.entries = pending

	return due
}

// Clear removes every pending task.
func (queue *Queue) Clear() {
	queue.mutex.Lock()
	queue.entries = nil
	queue.mutex.Unlock()
}

// Len returns the pending task count.
func (queue *Queue) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	return len(queue.entries)
}

// schedule stores one task and optionally replaces its key.
func (queue *Queue) schedule(key Key, deadline time.Time, run Run, replace bool) {
	if run == nil {
		return
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if replace && key != 0 {
		for index := range queue.entries {
			if queue.entries[index].key == key {
				queue.entries[index] = entry{key: key, deadline: deadline, run: run}
				return
			}
		}
	}
	queue.entries = append(queue.entries, entry{key: key, deadline: deadline, run: run})
}
