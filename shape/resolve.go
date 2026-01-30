package shape

import "fmt"

func resolveFile(file *File) error {
	if file == nil {
		return nil
	}
	resolved := map[string]bool{}
	var resolveShape func(name string, stack map[string]bool) (*Type, error)
	resolveShape = func(name string, stack map[string]bool) (*Type, error) {
		if stack[name] {
			return nil, fmt.Errorf("shape parse error: recursive type reference: %s", name)
		}
		if resolved[name] {
			return file.ByName[name].Type, nil
		}
		sh, ok := file.ByName[name]
		if !ok {
			return nil, fmt.Errorf("shape parse error: unknown type: %s", name)
		}
		stack[name] = true
		t, err := resolveType(sh.Type, stack, resolveShape)
		if err != nil {
			return nil, err
		}
		sh.Type = t
		resolved[name] = true
		delete(stack, name)
		return t, nil
	}

	for name := range file.ByName {
		if _, err := resolveShape(name, map[string]bool{}); err != nil {
			return err
		}
	}
	return nil
}

func resolveType(t *Type, stack map[string]bool, resolveShape func(string, map[string]bool) (*Type, error)) (*Type, error) {
	if t == nil {
		return nil, nil
	}
	switch t.Kind {
	case KindArray:
		elem, err := resolveType(t.Elem, stack, resolveShape)
		if err != nil {
			return nil, err
		}
		t.Elem = elem
		return t, nil
	case KindObject:
		for i := range t.Fields {
			ft, err := resolveType(t.Fields[i].Type, stack, resolveShape)
			if err != nil {
				return nil, err
			}
			t.Fields[i].Type = ft
		}
		return t, nil
	case KindRef:
		if t.RefName == "" {
			return nil, fmt.Errorf("shape parse error: empty type reference")
		}
		return resolveShape(t.RefName, stack)
	case KindUnion:
		for i := range t.Options {
			opt, err := resolveType(t.Options[i], stack, resolveShape)
			if err != nil {
				return nil, err
			}
			t.Options[i] = opt
		}
		return t, nil
	default:
		return t, nil
	}
}
