// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/creachadair/pson/textpb"
)

var configs = []Config{
	{false, false, "@"},
	{false, true, "@"},
	{true, false, "@"},
	{true, true, "@"},
}

var sub = strings.NewReplacer("*", "\n")
var a2c = strings.NewReplacer("<", "{", ">", "}")

func ans(verbose, compact string) []string {
	v := sub.Replace(verbose)
	c := sub.Replace(compact)
	return []string{v, a2c.Replace(v), c, a2c.Replace(c)}
}

func TestTextFormatting(t *testing.T) {
	tests := []struct {
		input  string
		answer []string
	}{
		{"", ans("", "")},
		{"a<\n>", ans("a <>", "a <>")},
		{"a: 1", ans("a: 1", "a:1")},
		{"a:\n\t\"foo\"\n", ans(`a: "foo"`, `a:"foo"`)},
		{"a<n:1>", ans("a <*@n: 1*>", "a <n:1>")},
		{`a<n:1 s:"foo">a<n:2 s:"bar">`,
			ans(`a <*@n: 1*@s: "foo"*>*a <*@n: 2*@s: "bar"*>`, `a <n:1 s:"foo"> a <n:2 s:"bar">`)},
		{`a { b {} } a {} c:<>`,
			ans(`a <*@b <>*>*a <>*c <>`, `a <b <>> a <> c <>`)},
		{`a{b{c:1 c:2 c:3}d{e{f:0x3f}}}`,
			ans(`a <*@b <*@@c: 1*@@c: 2*@@c: 3*@>*@d <*@@e <*@@@f: 0x3f*@@>*@>*>`, `a <b <c:1 c:2 c:3> d <e <f:0x3f>>>`)},
		{`a:FOO a:BAR a:BAZ`,
			ans(`a: FOO*a: BAR*a: BAZ`, `a:FOO a:BAR a:BAZ`)},
	}
	for _, test := range tests {
		msg, err := textpb.ParseString(test.input)
		if err != nil {
			t.Fatalf("[BROKEN TEST] Parsing %q failed: %v", test.input, err)
		}
		for i, cfg := range configs {
			t.Logf("Config: %+v", cfg)
			var buf bytes.Buffer
			if err := cfg.Text(&buf, msg.Combine()); err != nil {
				t.Errorf("Text %#q: unexpected error: %v", test.input, err)
				continue
			}
			got, want := buf.String(), test.answer[i]
			if got != want {
				t.Errorf("Text %#q: got\n«%s»\nwant\n«%s»", test.input, got, want)
			}
		}
	}
}
