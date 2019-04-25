// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package textpb

import "fmt"

func (m Message) ToValue() (interface{}, error) {
	out := make(map[string]interface{})
	for _, f := range m {
		if len(f.Values) == 1 {
			v, err := f.Values[0].ToValue()
			if err != nil {
				return nil, err
			}
			out[f.Name] = v
			continue
		}
		var vals []interface{}
		for _, v := range f.Values {
			w, err := v.ToValue()
			if err != nil {
				return nil, err
			}
			vals = append(vals, w)
		}
		out[f.Name] = vals
	}
	return out, nil
}

func (v *Value) ToValue() (interface{}, error) {
	if v.Msg != nil {
		return v.Msg.ToValue()
	}
	switch v.Type {
	case None:
		return nil, nil
	case Name, String, TypeName:
		return v.Text, nil
	case True:
		return true, nil
	case False:
		return false, nil
	case Number:
		if fix, err := v.Fixed(); err == nil {
			return fix, nil
		} else if fp, err := v.Number(); err == nil {
			return fp, nil
		}
		return nil, fmt.Errorf("inconvertible number: %q", v.Text)
	default:
		return nil, fmt.Errorf("invalid value type: %v", v.Type)
	}
	// unreachable
}
