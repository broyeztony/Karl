package ast

import (
	"encoding/json"
	"fmt"
)

// FormatJSON returns a pretty-printed JSON view of the AST.
func FormatJSON(node Node) (string, error) {
	value := toJSON(node)
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func toJSON(node Node) interface{} {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *Program:
		return map[string]interface{}{
			"type":       "Program",
			"statements": statementsToJSON(n.Statements),
		}
	case *LetStatement:
		return map[string]interface{}{
			"type":  "LetStatement",
			"name":  toJSON(n.Name),
			"value": toJSON(n.Value),
		}
	case *ExpressionStatement:
		return map[string]interface{}{
			"type":       "ExpressionStatement",
			"expression": toJSON(n.Expression),
		}
	case *Identifier:
		return map[string]interface{}{
			"type":  "Identifier",
			"value": n.Value,
		}
	case *Placeholder:
		return map[string]interface{}{
			"type": "Placeholder",
		}
	case *IntegerLiteral:
		return map[string]interface{}{
			"type":  "IntegerLiteral",
			"value": n.Value,
		}
	case *FloatLiteral:
		return map[string]interface{}{
			"type":  "FloatLiteral",
			"value": n.Value,
		}
	case *StringLiteral:
		return map[string]interface{}{
			"type":  "StringLiteral",
			"value": n.Value,
		}
	case *CharLiteral:
		return map[string]interface{}{
			"type":  "CharLiteral",
			"value": n.Value,
		}
	case *BooleanLiteral:
		return map[string]interface{}{
			"type":  "BooleanLiteral",
			"value": n.Value,
		}
	case *NullLiteral:
		return map[string]interface{}{
			"type": "NullLiteral",
		}
	case *UnitLiteral:
		return map[string]interface{}{
			"type": "UnitLiteral",
		}
	case *PrefixExpression:
		return map[string]interface{}{
			"type":     "PrefixExpression",
			"operator": n.Operator,
			"right":    toJSON(n.Right),
		}
	case *InfixExpression:
		return map[string]interface{}{
			"type":     "InfixExpression",
			"operator": n.Operator,
			"left":     toJSON(n.Left),
			"right":    toJSON(n.Right),
		}
	case *AssignExpression:
		return map[string]interface{}{
			"type":     "AssignExpression",
			"operator": n.Operator,
			"left":     toJSON(n.Left),
			"right":    toJSON(n.Right),
		}
	case *PostfixExpression:
		return map[string]interface{}{
			"type":     "PostfixExpression",
			"operator": n.Operator,
			"left":     toJSON(n.Left),
		}
	case *AwaitExpression:
		return map[string]interface{}{
			"type":  "WaitExpression",
			"value": toJSON(n.Value),
		}
	case *ImportExpression:
		return map[string]interface{}{
			"type": "ImportExpression",
			"path": toJSON(n.Path),
		}
	case *IfExpression:
		return map[string]interface{}{
			"type":        "IfExpression",
			"condition":   toJSON(n.Condition),
			"consequence": toJSON(n.Consequence),
			"alternative": toJSON(n.Alternative),
		}
	case *BlockExpression:
		return map[string]interface{}{
			"type":       "BlockExpression",
			"statements": statementsToJSON(n.Statements),
		}
	case *MatchExpression:
		arms := make([]interface{}, 0, len(n.Arms))
		for _, arm := range n.Arms {
			arms = append(arms, map[string]interface{}{
				"pattern": toJSON(arm.Pattern),
				"guard":   toJSON(arm.Guard),
				"body":    toJSON(arm.Body),
			})
		}
		return map[string]interface{}{
			"type":  "MatchExpression",
			"value": toJSON(n.Value),
			"arms":  arms,
		}
	case *ForExpression:
		bindings := make([]interface{}, 0, len(n.Bindings))
		for _, b := range n.Bindings {
			bindings = append(bindings, map[string]interface{}{
				"pattern": toJSON(b.Pattern),
				"value":   toJSON(b.Value),
			})
		}
		return map[string]interface{}{
			"type":      "ForExpression",
			"condition": toJSON(n.Condition),
			"bindings":  bindings,
			"body":      toJSON(n.Body),
			"then":      toJSON(n.Then),
		}
	case *LambdaExpression:
		return map[string]interface{}{
			"type":   "LambdaExpression",
			"params": patternsToJSON(n.Params),
			"body":   toJSON(n.Body),
		}
	case *CallExpression:
		return map[string]interface{}{
			"type":     "CallExpression",
			"function": toJSON(n.Function),
			"args":     expressionsToJSON(n.Arguments),
		}
	case *RecoverExpression:
		return map[string]interface{}{
			"type":     "RecoverExpression",
			"target":   toJSON(n.Target),
			"fallback": toJSON(n.Fallback),
		}
	case *MemberExpression:
		return map[string]interface{}{
			"type":     "MemberExpression",
			"object":   toJSON(n.Object),
			"property": toJSON(n.Property),
		}
	case *IndexExpression:
		return map[string]interface{}{
			"type":  "IndexExpression",
			"left":  toJSON(n.Left),
			"index": toJSON(n.Index),
		}
	case *SliceExpression:
		return map[string]interface{}{
			"type":  "SliceExpression",
			"left":  toJSON(n.Left),
			"start": toJSON(n.Start),
			"end":   toJSON(n.End),
		}
	case *ArrayLiteral:
		return map[string]interface{}{
			"type":     "ArrayLiteral",
			"elements": expressionsToJSON(n.Elements),
		}
	case *ObjectLiteral:
		entries := make([]interface{}, 0, len(n.Entries))
		for _, entry := range n.Entries {
			entries = append(entries, map[string]interface{}{
				"key":       entry.Key,
				"value":     toJSON(entry.Value),
				"shorthand": entry.Shorthand,
				"spread":    entry.Spread,
			})
		}
		return map[string]interface{}{
			"type":    "ObjectLiteral",
			"entries": entries,
		}
	case *StructInitExpression:
		return map[string]interface{}{
			"type":     "StructInitExpression",
			"typeName": toJSON(n.TypeName),
			"value":    toJSON(n.Value),
		}
	case *RangeExpression:
		return map[string]interface{}{
			"type":  "RangeExpression",
			"start": toJSON(n.Start),
			"end":   toJSON(n.End),
			"step":  toJSON(n.Step),
		}
	case *QueryExpression:
		return map[string]interface{}{
			"type":    "QueryExpression",
			"var":     toJSON(n.Var),
			"source":  toJSON(n.Source),
			"where":   expressionsToJSON(n.Where),
			"orderBy": toJSON(n.OrderBy),
			"select":  toJSON(n.Select),
		}
	case *RaceExpression:
		return map[string]interface{}{
			"type":  "RaceExpression",
			"tasks": expressionsToJSON(n.Tasks),
		}
	case *SpawnExpression:
		return map[string]interface{}{
			"type":  "SpawnExpression",
			"task":  toJSON(n.Task),
			"group": expressionsToJSON(n.Group),
		}
	case *BreakExpression:
		return map[string]interface{}{
			"type":  "BreakExpression",
			"value": toJSON(n.Value),
		}
	case *ContinueExpression:
		return map[string]interface{}{
			"type": "ContinueExpression",
		}
	case *WildcardPattern:
		return map[string]interface{}{
			"type": "WildcardPattern",
		}
	case *RangePattern:
		return map[string]interface{}{
			"type":  "RangePattern",
			"start": toJSON(n.Start),
			"end":   toJSON(n.End),
		}
	case *ObjectPattern:
		entries := make([]interface{}, 0, len(n.Entries))
		for _, entry := range n.Entries {
			entries = append(entries, map[string]interface{}{
				"key":     entry.Key,
				"pattern": toJSON(entry.Pattern),
			})
		}
		return map[string]interface{}{
			"type":    "ObjectPattern",
			"entries": entries,
		}
	case *ArrayPattern:
		return map[string]interface{}{
			"type":     "ArrayPattern",
			"elements": patternsToJSON(n.Elements),
			"rest":     toJSON(n.Rest),
		}
	case *TuplePattern:
		return map[string]interface{}{
			"type":     "TuplePattern",
			"elements": patternsToJSON(n.Elements),
		}
	case *CallPattern:
		return map[string]interface{}{
			"type": "CallPattern",
			"name": toJSON(n.Name),
			"args": patternsToJSON(n.Args),
		}
	default:
		return map[string]interface{}{
			"type":    "Unknown",
			"go_type": fmt.Sprintf("%T", node),
		}
	}
}

func nodesToJSON(nodes []Node) []interface{} {
	list := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		list = append(list, toJSON(n))
	}
	return list
}

func statementsToJSON(nodes []Statement) []interface{} {
	list := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		list = append(list, toJSON(n))
	}
	return list
}

func expressionsToJSON(nodes []Expression) []interface{} {
	list := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		list = append(list, toJSON(n))
	}
	return list
}

func patternsToJSON(nodes []Pattern) []interface{} {
	list := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		list = append(list, toJSON(n))
	}
	return list
}
