package db

import (
	"errors"
	"testing"
)

type trackingObserver struct {
	NoopObserver
	creating int
	created  int
	updating int
	updated  int
	deleting int
	deleted  int
	blockOn  string // "creating" | "updating" | "deleting" → return error
}

func (o *trackingObserver) Creating(data map[string]any) error {
	o.creating++
	if o.blockOn == "creating" {
		return errors.New("blocked")
	}
	return nil
}
func (o *trackingObserver) Created(data map[string]any)  { o.created++ }
func (o *trackingObserver) Updating(data map[string]any) error {
	o.updating++
	if o.blockOn == "updating" {
		return errors.New("blocked")
	}
	return nil
}
func (o *trackingObserver) Updated(data map[string]any) { o.updated++ }
func (o *trackingObserver) Deleting(id any) error {
	o.deleting++
	if o.blockOn == "deleting" {
		return errors.New("blocked")
	}
	return nil
}
func (o *trackingObserver) Deleted(id any) { o.deleted++ }

func TestNoopObserver(t *testing.T) {
	var o NoopObserver
	if err := o.Creating(nil); err != nil {
		t.Fatalf("NoopObserver.Creating: %v", err)
	}
	if err := o.Updating(nil); err != nil {
		t.Fatalf("NoopObserver.Updating: %v", err)
	}
	if err := o.Deleting(nil); err != nil {
		t.Fatalf("NoopObserver.Deleting: %v", err)
	}
	o.Created(nil)
	o.Updated(nil)
	o.Deleted(nil)
}

func TestObserve_Register(t *testing.T) {
	m := NewModel(nil, "users")
	obs := &trackingObserver{}
	m.Observe(obs)
	if len(m.observers) != 1 {
		t.Fatalf("expected 1 observer, got %d", len(m.observers))
	}
}

func TestObserve_MultipleObservers(t *testing.T) {
	m := NewModel(nil, "users")
	o1 := &trackingObserver{}
	o2 := &trackingObserver{}
	m.Observe(o1)
	m.Observe(o2)
	if len(m.observers) != 2 {
		t.Fatalf("expected 2 observers, got %d", len(m.observers))
	}
}

func TestObserveDeleting_Blocked(t *testing.T) {
	obs := &trackingObserver{blockOn: "deleting"}
	m := NewModel(nil, "users")
	m.Observe(obs)

	// ObservedDestroy calls Deleting hook — should return error before hitting DB
	err := m.ObservedDestroy(1)
	if err == nil {
		t.Fatal("expected error when Deleting hook blocks")
	}
	if obs.deleting != 1 {
		t.Fatalf("expected deleting=1, got %d", obs.deleting)
	}
	if obs.deleted != 0 {
		t.Fatal("Deleted hook should not be called when Deleting blocks")
	}
}

func TestObserveCreating_Blocked(t *testing.T) {
	obs := &trackingObserver{blockOn: "creating"}
	m := NewModel(nil, "users")
	m.Observe(obs)

	_, err := m.ObservedCreate(map[string]any{"name": "Alice"})
	if err == nil {
		t.Fatal("expected error when Creating hook blocks")
	}
	if obs.creating != 1 {
		t.Fatalf("expected creating=1, got %d", obs.creating)
	}
	if obs.created != 0 {
		t.Fatal("Created hook should not be called when Creating blocks")
	}
}

func TestObserveUpdating_Blocked(t *testing.T) {
	obs := &trackingObserver{blockOn: "updating"}
	m := NewModel(nil, "users")
	m.Observe(obs)

	err := m.ObservedSave(1, map[string]any{"name": "Bob"})
	if err == nil {
		t.Fatal("expected error when Updating hook blocks")
	}
	if obs.updating != 1 {
		t.Fatalf("expected updating=1, got %d", obs.updating)
	}
	if obs.updated != 0 {
		t.Fatal("Updated hook should not be called when Updating blocks")
	}
}
