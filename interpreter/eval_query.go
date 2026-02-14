package interpreter

import (
	"karl/ast"
	"sort"
)

type queryRow struct {
	item Value
	key  Value
}

func (e *Evaluator) evalQueryExpression(node *ast.QueryExpression, env *Environment) (Value, *Signal, error) {
	sourceVal, sig, err := e.Eval(node.Source, env)
	if err != nil || sig != nil {
		return sourceVal, sig, err
	}
	source, ok := sourceVal.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "query source must be array"}
	}

	rows := []queryRow{}

	for _, item := range source.Elements {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, item)

		keep := true
		for _, whereExpr := range node.Where {
			val, sig, err := e.Eval(whereExpr, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			b, ok := val.(*Boolean)
			if !ok {
				return nil, nil, &RuntimeError{Message: "query where must be bool"}
			}
			if !b.Value {
				keep = false
				break
			}
		}

		if !keep {
			continue
		}

		var key Value
		if node.OrderBy != nil {
			val, sig, err := e.Eval(node.OrderBy, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			key = val
		}
		rows = append(rows, queryRow{item: item, key: key})
	}

	if node.OrderBy != nil {
		if err := sortRows(rows); err != nil {
			return nil, nil, err
		}
	}

	results := []Value{}
	for _, r := range rows {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, r.item)
		val, sig, err := e.Eval(node.Select, rowEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		results = append(results, val)
	}
	return &Array{Elements: results}, nil, nil
}

func sortRows(rows []queryRow) error {
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.key == nil {
			return &RuntimeError{Message: "orderby requires comparable key"}
		}
	}

	switch rows[0].key.(type) {
	case *Integer, *Float, *String:
	default:
		return &RuntimeError{Message: "orderby key must be int, float, or string"}
	}
	firstType := rows[0].key.Type()
	for _, row := range rows {
		if row.key.Type() != firstType {
			return &RuntimeError{Message: "orderby keys must be the same type"}
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return compareForSort(rows[i].key, rows[j].key)
	})
	return nil
}

func compareForSort(left, right Value) bool {
	switch l := left.(type) {
	case *Integer:
		return l.Value < right.(*Integer).Value
	case *Float:
		return l.Value < right.(*Float).Value
	case *String:
		return l.Value < right.(*String).Value
	default:
		return false
	}
}
