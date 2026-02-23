package spreadsheet

import "testing"

func findUpdateByID(updates []UpdateResponse, id string) *UpdateResponse {
	for i := range updates {
		if updates[i].ID == id {
			return &updates[i]
		}
	}
	return nil
}

func TestRuntimeUpdateCellPropagatesDependents(t *testing.T) {
	rt := NewRuntime()
	rt.Clear()

	if _, err := rt.UpdateCell("A1", "10"); err != nil {
		t.Fatalf("set A1 failed: %v", err)
	}
	if _, err := rt.UpdateCell("B1", "= A1 * 2"); err != nil {
		t.Fatalf("set B1 failed: %v", err)
	}

	updates, err := rt.UpdateCell("A1", "7")
	if err != nil {
		t.Fatalf("update A1 failed: %v", err)
	}

	a1 := findUpdateByID(updates, "A1")
	if a1 == nil {
		t.Fatalf("missing A1 update")
	}
	if a1.Display != "7" {
		t.Fatalf("unexpected A1 display: %q", a1.Display)
	}

	b1 := findUpdateByID(updates, "B1")
	if b1 == nil {
		t.Fatalf("missing B1 update")
	}
	if b1.Display != "14" {
		t.Fatalf("unexpected B1 display: %q", b1.Display)
	}
}

func TestRuntimeLoadExampleAndSnapshot(t *testing.T) {
	rt := NewRuntime()
	rt.Clear()

	if got := len(rt.Snapshot()); got != 0 {
		t.Fatalf("expected empty snapshot after clear, got %d cells", got)
	}

	rt.LoadExample("intro")
	if got := len(rt.Snapshot()); got == 0 {
		t.Fatalf("expected populated snapshot for intro example")
	}
}
