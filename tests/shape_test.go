package tests

import (
	"path/filepath"
	"testing"

	"karl/ast"
)

func TestExampleShapes(t *testing.T) {
	t.Run("03_loop_for", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "loop_for.k"))
		requireCountAtLeast(t, program, "ForExpression", 2, func(n ast.Node) bool {
			_, ok := n.(*ast.ForExpression)
			return ok
		})
		requireCountAtLeast(t, program, "BreakExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.BreakExpression)
			return ok
		})
		requireCountAtLeast(t, program, "ContinueExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.ContinueExpression)
			return ok
		})
	})

	t.Run("05_concurrency_basic", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "concurrency", "basic.k"))
		requireCountAtLeast(t, program, "SpawnExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.SpawnExpression)
			return ok
		})
		requireCountAtLeast(t, program, "RaceExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.RaceExpression)
			return ok
		})
	})

	t.Run("08_query_basic", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "query_basic.k"))
		requireCountAtLeast(t, program, "QueryExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.QueryExpression)
			return ok
		})
	})

	t.Run("09_ranges_slices", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "ranges_slices.k"))
		requireCountAtLeast(t, program, "RangeExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.RangeExpression)
			return ok
		})
	})
}

func TestExerciseShapes(t *testing.T) {
	t.Run("01_channels_and_cancel", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "concurrency", "channels_and_cancel.k"))
		requireCountAtLeast(t, program, "SpawnExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.SpawnExpression)
			return ok
		})
	})

	t.Run("05_loop_for", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "loop_for.k"))
		requireCountAtLeast(t, program, "ForExpression", 2, func(n ast.Node) bool {
			_, ok := n.(*ast.ForExpression)
			return ok
		})
	})

	t.Run("08_loop_for_break", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "features", "loop_for.k"))
		requireCountAtLeast(t, program, "BreakExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.BreakExpression)
			return ok
		})
	})
}
