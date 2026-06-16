package state

import (
	"testing"
)

func TestAddDedupAndGet(t *testing.T) {
	s := &State{Todos: map[string][]string{}}
	s.Add("2026-06-16", "alpha")
	s.Add("2026-06-16", "alpha") // duplicate ignored
	s.Add("2026-06-16", "beta")

	got := s.Get("2026-06-16")
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("got %v, want [alpha beta]", got)
	}
}

func TestMostRecentBefore(t *testing.T) {
	s := &State{Todos: map[string][]string{
		"2026-06-10": {"a"},
		"2026-06-14": {"b"},
		"2026-06-16": {"c"}, // == query date, excluded
		"2026-06-15": {},    // empty, excluded
	}}

	d, ok := s.MostRecentBefore("2026-06-16")
	if !ok || d != "2026-06-14" {
		t.Fatalf("got (%q,%v), want (2026-06-14,true)", d, ok)
	}

	if _, ok := s.MostRecentBefore("2026-06-01"); ok {
		t.Fatalf("expected no date before 2026-06-01")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	// Load on an empty repo returns empty (no .git needed for this unit test).
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	s.Add("2026-06-16", "x")
	if err := s.Save(dir); err != nil {
		t.Fatal(err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if names := got.Get("2026-06-16"); len(names) != 1 || names[0] != "x" {
		t.Fatalf("round-trip lost data: %v", names)
	}
}
