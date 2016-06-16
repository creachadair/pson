package textpb

import (
	"fmt"
	"strings"
	"testing"
)

func pathValue(m Message, path string) (*Value, error) {
	if path == "" {
		return nil, nil
	}
	cur := m
	var val *Value
	for _, name := range strings.Split(path, ".") {
		var found *Field
		for _, f := range cur {
			if f.Name == name {
				found = f
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("path %q not found", path)
		} else if len(found.Values) == 0 {
			return nil, fmt.Errorf("%q has no values", name)
		}
		val = found.Values[0]
		cur = val.Msg
	}
	if val != nil {
		return val, nil
	}
	return nil, fmt.Errorf("no value at path %q", path)
}

func TestParse(t *testing.T) {
	tests := []struct {
		input, path, want string
	}{
		{"", "", ""},
		{"a:1", "a", "1"},
		{"a:true", "a", "true"},
		{`a:"foo" b:false`, "b", "false"},
		{`a < in: true >`, "a.in", "true"},
		{`a:"b" "c" "d"`, "a", "bcd"},
		{`a:'b' 'c' "d"`, "a", "bcd"},
		{`a < n:1 s:"two" > b { n:2 s:false}`, "a.s", "two"},
		{`a: < b <> c: false d < [x]: { y:1 } >>`, "a.d.x.y", "1"},
		{`a:1, b:2; c < d < e:3, > >;`, "c.d.e", "3"},
		{`# Pearls and swine
bereft: "of" ' me'

long_and_weary: < my: road has: been > # I was lost

in: 'the cities'
  alone <
    in: 'the hills' # no sorrow
  >

# or pity for leaving I feel`, "long_and_weary.has", "been"},
	}
	for _, test := range tests {
		got, err := Parse(strings.NewReader(test.input))
		if err != nil {
			t.Errorf("Parse %q failed: %v", test.input, err)
			continue
		}
		v, err := pathValue(got, test.path)
		if err != nil {
			t.Errorf("Lookup failed: %v", err)
		} else if v != nil && v.Text != test.want {
			t.Errorf("Value of %q: got %q, want %q", test.path, v.Text, test.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []string{
		// Missing field name
		"<>", ":true",

		// Bogus field name
		"1:true", "true:'false'",

		// Various unbalanced things...
		"a <", "a: >", "a {", "a: }", "a: '", `a: "`,

		// Type names require message values
		"[a/b/c]: wrong",
	}
	for _, test := range tests {
		got, err := Parse(strings.NewReader(test))
		if err == nil {
			t.Errorf("Parse %q: got %+v, wanted error", test, got)
			continue
		}
		t.Logf("Parse %q OK: got error %v", test, err)
	}
}
