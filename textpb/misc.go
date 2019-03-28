// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package textpb

// This file adds split/combine and other utility code.

import "sort"

// Camel recursively renames each field of m in-place, converting "snake_case"
// names to "camelCase".
func (m Message) ToCamel() {
	for _, f := range m {
		f.toCamel()
	}
}

func (f *Field) toCamel() {
	f.Name = SnakeToCamel(f.Name)
	for _, v := range f.Values {
		v.toCamel()
	}
}

func (v *Value) toCamel() {
	if v.Msg != nil {
		v.Msg.ToCamel()
	}
}

// Combine returns a copy of m in which each field name occurs exactly once,
// with all the values assigned to that field name.  This process is applied
// recursively to nested messages.
func (m Message) Combine() Message {
	names := make(map[string]*Field)
	for _, field := range m {
		of := names[field.Name]
		if of == nil {
			of = &Field{Name: field.Name}
			names[field.Name] = of
		}
		for _, v := range field.Values {
			of.Values = append(of.Values, v.combine())
		}
	}
	out := Message{}
	for _, field := range names {
		out = append(out, field)
	}
	sort.Sort(out)
	return out
}

// RSplit recursively partitions m into multiple messages with the property
// that each field of each resulting message has at most one value.
func (m Message) RSplit() []Message { return m.Combine().split(true) }

// Split partitions m into multiple messages with the property that each field
// of each resulting message has at most one value.
func (m Message) Split() []Message { return m.Combine().split(false) }

func (m Message) split(recur bool) []Message {
	var all [][]*Field // the results of partitioning all the fields
	for _, f := range m {
		if fs := f.split(recur); len(fs) > 0 {
			all = append(all, fs)
		}
	}

	// Accumulates the result messages.
	var result []Message

	// A slice of indexes into the field sets returned by expansion.
	// Each idx[i] is in the range [0,len(all[i])) and denotes the
	// position of the next field value to be consumed.
	idx := make([]int, len(all))

	done := false
	for !done {
		// Increment the index vector.
		for i, x := range idx {
			idx[i] = (x + 1) % len(all[i])
			if idx[i] != 0 {
				break
			}
			done = i+1 == len(idx)
		}

		// Copy the index values into the result.
		next := make(Message, len(idx))
		for i, x := range idx {
			next[i] = all[i][x]
		}
		result = append(result, next)
	}

	return result
}

func (f *Field) split(recur bool) []*Field {
	if len(f.Values) == 0 {
		return nil
	}
	var fs []*Field
	for _, v := range f.Values {
		for _, vs := range v.split(recur) {
			fs = append(fs, &Field{
				Name:   f.Name,
				Values: []*Value{vs},
			})
		}
	}
	return fs
}

func (v *Value) combine() *Value {
	if v.Msg == nil {
		return v
	}
	return &Value{Msg: v.Msg.Combine()}
}

func (v *Value) split(recur bool) []*Value {
	if v.Msg == nil || !recur {
		return []*Value{v}
	}
	var vs []*Value
	for _, msg := range v.Msg.split(recur) {
		vs = append(vs, &Value{Msg: msg})
	}
	return vs
}
