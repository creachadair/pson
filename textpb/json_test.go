// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package textpb

import "testing"

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", ""},
		{"_", ""},
		{"_alpha_", "alpha"},
		{"__init__", "init"},
		{"ALPHA", "alpha"},
		{"Alpha", "alpha"},
		{"INT_MAX", "intMax"},
		{"all_your_base", "allYourBase"},
		{"__private_parts", "privateParts"},
		{"let_FREEDOM_RinG", "letFreedomRing"},
	}
	for _, test := range tests {
		got := SnakeToCamel(test.input)
		if got != test.want {
			t.Errorf("SnakeToCamel(%q): got %q, want %q", test.input, got, test.want)
		}
	}
}
