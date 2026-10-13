// Package queue provides a bounded async channel for usage records.
// Overflow drops the incoming (newest) record and increments a counter.
// Already queued records retain FIFO order; Enqueue never waits for capacity.
// Queue errors never propagate to the request path.
package queue

import (
	"sync"
	"sync/atomic"
)

const DefaultCapacity = 2048

// Record is the internal queue entry — a pre-serialized JSON line ready to append.
type Record []byte

// Queue is a bounded channel of serialized usage records.
type Queue struct {
	ch       chan Record
	dropped  atomic.Int64
	shutdown chan struct{}
	mu       sync.Mutex
	closed   bool
}

// New creates a queue with the given capacity.
func New(capacity int) *Queue {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Queue{
		ch:       make(chan Record, capacity),
		shutdown: make(chan struct{}),
	}
}

// Enqueue adds a record to the queue non-blocking.
// If the queue is full or closed, it drops the incoming record and increments
// the dropped counter without evicting existing entries.
func (q *Queue) Enqueue(r Record) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		q.dropped.Add(1)
		return
	}
	select {
	case q.ch <- r:
	default:
		q.dropped.Add(1)
	}
}

// Consume returns the channel for reading records.
func (q *Queue) Consume() <-chan Record {
	return q.ch
}

// DroppedCount returns the cumulative number of dropped records.
func (q *Queue) DroppedCount() int64 {
	return q.dropped.Load()
}

// Close signals shutdown; the consumer drains remaining records.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	close(q.shutdown)
}

// ShutdownCh returns the shutdown channel.
func (q *Queue) ShutdownCh() <-chan struct{} {
	return q.shutdown
}
