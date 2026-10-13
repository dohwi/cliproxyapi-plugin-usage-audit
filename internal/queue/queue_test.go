package queue_test

import (
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
)

func TestOverflowDropsIncomingPreservingFIFO(t *testing.T) {
	q := queue.New(2)
	defer q.Close()
	q.Enqueue(queue.Record("first"))
	q.Enqueue(queue.Record("second"))
	q.Enqueue(queue.Record("dropped-third"))
	q.Enqueue(queue.Record("dropped-fourth"))
	if got := q.DroppedCount(); got != 2 {
		t.Fatalf("drops = %d, want 2", got)
	}
	for _, want := range []string{"first", "second"} {
		select {
		case got := <-q.Consume():
			if string(got) != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("missing queued record")
		}
	}
	q.Enqueue(queue.Record("after-drain"))
	q.Close()
	q.Close()
	q.Enqueue(queue.Record("after-close"))
	if got := q.DroppedCount(); got != 3 {
		t.Fatalf("drops after close = %d, want 3", got)
	}
	select {
	case <-q.ShutdownCh():
	default:
		t.Fatal("shutdown not signaled")
	}
	select {
	case got := <-q.Consume():
		if string(got) != "after-drain" {
			t.Fatalf("pending close record = %q", got)
		}
	default:
		t.Fatal("close discarded pending record")
	}
	select {
	case got := <-q.Consume():
		t.Fatalf("unexpected record %q", got)
	default:
	}
}
