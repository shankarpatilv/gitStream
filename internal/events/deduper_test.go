package events

import "testing"

func TestDeduperAcceptsFirstObservation(t *testing.T) {
	deduper := NewDeduper(1000)

	if !deduper.IsNew("event-1") {
		t.Fatal("first observation should be accepted")
	}
}

func TestDeduperRejectsSecondObservation(t *testing.T) {
	deduper := NewDeduper(1000)

	if !deduper.IsNew("event-1") {
		t.Fatal("first observation should be accepted")
	}
	if deduper.IsNew("event-1") {
		t.Fatal("second observation should be rejected")
	}
}

func TestDeduperBoundsMemoryToLimit(t *testing.T) {
	deduper := NewDeduper(2)

	deduper.IsNew("event-1")
	deduper.IsNew("event-2")
	deduper.IsNew("event-3")

	if len(deduper.seen) != 2 {
		t.Fatalf("len(seen) = %d, want 2", len(deduper.seen))
	}
	if len(deduper.order) != 2 {
		t.Fatalf("len(order) = %d, want 2", len(deduper.order))
	}
	if deduper.IsNew("event-1") != true {
		t.Fatal("evicted event should be accepted again")
	}
	if deduper.IsNew("event-3") != false {
		t.Fatal("recent event should still be rejected")
	}
}

func TestDeduperRejectsEmptyID(t *testing.T) {
	deduper := NewDeduper(1000)

	if deduper.IsNew("") {
		t.Fatal("empty id should be rejected")
	}
}
