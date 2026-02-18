package spreadsheet

import (
	"karl/interpreter"
	"testing"
)

func TestSimpleEvaluation(t *testing.T) {
	s := NewSheet()
	
	// Set A1 = 10
	if err := s.SetCell("A1", "10"); err != nil {
		t.Fatalf("Failed to set A1: %v", err)
	}
	
	cell := s.GetCell("A1")
	if val, ok := cell.Value.(*interpreter.Integer); !ok || val.Value != 10 {
		t.Errorf("Expected A1 to be 10, got %v", cell.Value)
	}
}

func TestDependencyPropagation(t *testing.T) {
	s := NewSheet()
	
	// A1 = 10
	s.SetCell("A1", "10")
	
	// B1 = A1 * 2
	if err := s.SetCell("B1", "= A1 * 2"); err != nil {
		t.Fatalf("Failed to set B1: %v", err)
	}
	
	// Check B1
	b1 := s.GetCell("B1")
	if b1.Error != nil {
		t.Fatalf("B1 has error: %v", b1.Error)
	}
	
	if val, ok := b1.Value.(*interpreter.Integer); !ok || val.Value != 20 {
		t.Errorf("Expected B1 to be 20, got %v", b1.Value)
	}
	
	// Update A1 = 5
	s.SetCell("A1", "5")
	
	// Check B1 again (should be 10)
	if val, ok := b1.Value.(*interpreter.Integer); !ok || val.Value != 10 {
		t.Errorf("Expected B1 to update to 10, got %v", b1.Value)
	}
}

func TestChainedDependencies(t *testing.T) {
	s := NewSheet()
	
	s.SetCell("A1", "1")
	s.SetCell("B1", "= A1 + 1")
	s.SetCell("C1", "= B1 * 2")
	
	c1 := s.GetCell("C1")
	// Expected: (1+1)*2 = 4
	if val, ok := c1.Value.(*interpreter.Integer); !ok || val.Value != 4 {
		t.Errorf("Expected C1 to be 4, got %v", c1.Value)
	}
	
	// Update A1 = 2 -> B1 = 3 -> C1 = 6
	s.SetCell("A1", "2")
	
	if val, ok := c1.Value.(*interpreter.Integer); !ok || val.Value != 6 {
		t.Errorf("Expected C1 to update to 6, got %v", c1.Value)
	}
}
