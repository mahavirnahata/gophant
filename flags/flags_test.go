package flags

import "testing"

func setup() { SetStore(NewMemoryStore()) }

func TestEnableDisable(t *testing.T) {
	setup()
	Enable("feature-a")
	if !IsEnabled("feature-a") {
		t.Fatal("expected feature-a to be enabled")
	}
	Disable("feature-a")
	if IsEnabled("feature-a") {
		t.Fatal("expected feature-a to be disabled")
	}
}

func TestDefaultDisabled(t *testing.T) {
	setup()
	if IsEnabled("unknown-flag") {
		t.Fatal("unknown flag should be disabled by default")
	}
}

func TestToggle(t *testing.T) {
	setup()
	Toggle("feature-b")
	if !IsEnabled("feature-b") {
		t.Fatal("expected enabled after first toggle")
	}
	Toggle("feature-b")
	if IsEnabled("feature-b") {
		t.Fatal("expected disabled after second toggle")
	}
}

func TestAll(t *testing.T) {
	setup()
	Enable("f1")
	Enable("f2")
	Disable("f1")
	all := All()
	if all["f1"] {
		t.Fatal("f1 should be false")
	}
	if !all["f2"] {
		t.Fatal("f2 should be true")
	}
}

func TestWhen_Enabled(t *testing.T) {
	setup()
	Enable("run-fn")
	ran := false
	When("run-fn", func() { ran = true })
	if !ran {
		t.Fatal("When should call fn when feature is enabled")
	}
}

func TestWhen_Disabled(t *testing.T) {
	setup()
	ran := false
	When("not-enabled", func() { ran = true })
	if ran {
		t.Fatal("When should not call fn when feature is disabled")
	}
}

func TestSetStore(t *testing.T) {
	s := NewMemoryStore()
	s.Enable("custom")
	SetStore(s)
	if !IsEnabled("custom") {
		t.Fatal("custom store not applied")
	}
	setup() // restore default
}

func TestMemoryStore_All_IsCopy(t *testing.T) {
	s := NewMemoryStore()
	s.Enable("x")
	all := s.All()
	all["x"] = false // mutate the copy
	if !s.IsEnabled("x") {
		t.Fatal("All() should return a copy, not a reference")
	}
}
