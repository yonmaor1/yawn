package config

import "testing"

func TestRoundTrip(t *testing.T) {
	in := &Task{
		Directory: "/Users/you/git/foo",
		Status:    StatusPending,
		Parent:    "done",
		RC:        []string{"source .venv/bin/activate"},
		Cleanup:   []string{"deactivate"},
	}
	b, err := in.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	out, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if out.Directory != in.Directory || out.Status != in.Status || out.Parent != in.Parent {
		t.Fatalf("scalar mismatch: %+v vs %+v", out, in)
	}
	if len(out.RC) != 1 || out.RC[0] != "source .venv/bin/activate" {
		t.Fatalf("rc mismatch: %v", out.RC)
	}
	if len(out.Cleanup) != 1 || out.Cleanup[0] != "deactivate" {
		t.Fatalf("cleanup mismatch: %v", out.Cleanup)
	}
}

func TestDefault(t *testing.T) {
	d := Default("myparent")
	if d.Status != StatusPending || d.Parent != "myparent" {
		t.Fatalf("unexpected default: %+v", d)
	}
}
