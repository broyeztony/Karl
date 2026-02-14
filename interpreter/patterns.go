package interpreter

import (
	"fmt"
	"karl/ast"
)

func matchPattern(pattern ast.Pattern, value Value, env *Environment) (bool, error) {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return true, nil
	case *ast.Identifier:
		env.Define(p.Value, value)
		return true, nil
	case *ast.IntegerLiteral:
		v, ok := value.(*Integer)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.FloatLiteral:
		v, ok := value.(*Float)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.StringLiteral:
		v, ok := value.(*String)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.CharLiteral:
		v, ok := value.(*Char)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.BooleanLiteral:
		v, ok := value.(*Boolean)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.NullLiteral:
		_, ok := value.(*Null)
		return ok, nil
	case *ast.RangePattern:
		return matchRangePattern(p, value)
	case *ast.ObjectPattern:
		return matchObjectPattern(p, value, env)
	case *ast.ArrayPattern:
		return matchArrayPattern(p, value, env)
	case *ast.TuplePattern:
		return matchTuplePattern(p, value, env)
	case *ast.CallPattern:
		return matchCallPattern(p, value, env)
	default:
		return false, fmt.Errorf("unsupported pattern type: %T", pattern)
	}
}
