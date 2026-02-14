package interpreter

import (
	"fmt"
	"karl/ast"
)

func matchRangePattern(p *ast.RangePattern, value Value) (bool, error) {
	start, err := patternLiteralValue(p.Start)
	if err != nil {
		return false, err
	}
	end, err := patternLiteralValue(p.End)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case *Integer:
		s, sok := start.(*Integer)
		e, eok := end.(*Integer)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	case *Float:
		s, sok := start.(*Float)
		e, eok := end.(*Float)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	case *Char:
		s, sok := start.(*Char)
		e, eok := end.(*Char)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	default:
		return false, nil
	}
}

func patternLiteralValue(pattern ast.Pattern) (Value, error) {
	switch p := pattern.(type) {
	case *ast.IntegerLiteral:
		return &Integer{Value: p.Value}, nil
	case *ast.FloatLiteral:
		return &Float{Value: p.Value}, nil
	case *ast.StringLiteral:
		return &String{Value: p.Value}, nil
	case *ast.CharLiteral:
		return &Char{Value: p.Value}, nil
	default:
		return nil, fmt.Errorf("range patterns require literal endpoints")
	}
}
