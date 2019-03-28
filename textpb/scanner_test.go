// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package textpb

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func scan(s *Scanner, f func(Token, string)) error {
	for s.Next() {
		f(s.Token(), s.Text())
	}
	return s.Err()
}

func TestScanner(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		{"", nil},
		{`a:true,b<>;`, []Token{Name, Colon, True, Comma, Name, LeftA, RightA, Semi}},
		{"# A comment!\nwhat: is your: quest # Another comment\n'eh?'", []Token{
			Name, Colon, Name, Name, Colon, Name, String,
		}},
		{`"mültípàss" 'lεveρbøt' KOOL`, []Token{String, String, Name}},
		{`kind: CALL`, []Token{Name, Colon, Name}},
		{`[grok.proto.Foo] { value: 27 weight: .2 }`, []Token{
			TypeName, LeftC, Name, Colon, Number, Name, Colon, Number, RightC,
		}},
		{`1 2. .3 -.4 5e16 -6e+9 .70E-1 88.81 11f -.5e-2f`, []Token{
			Number, Number, Number, Number, Number, Number, Number, Number, Number, Number,
		}},
		{`decorations < outline:true source_text:false > ticket: "bogus"`, []Token{
			Name, LeftA, Name, Colon, True, Name, Colon, False, RightA, Name, Colon, String,
		}},
	}
	for _, test := range tests {
		s := NewScanner(strings.NewReader(test.input))
		var got []Token
		if err := scan(s, func(tok Token, text string) {
			got = append(got, tok)
			t.Logf("Got token [%v] %#q", tok, text)
		}); err != io.EOF {
			t.Errorf("Scan: got error %v, want %v", err, io.EOF)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Scan: got tokens %+v, want %+v", got, test.want)
		}
	}
}

func TestScanErrors(t *testing.T) {
	tests := []string{
		`"bad string`,
		"'bad\nstring'",
		`'whatcha gonna do`,
		`[whatcha gonna do]`,
		`when/they^come%for&you`,
		`?`,
		`-`, `.`, `.-9`, `2^&#$^@#$`,
	}
	for _, test := range tests {
		s := NewScanner(strings.NewReader(test))
		if err := scan(s, func(tok Token, text string) {
			t.Logf("Got token [%v] %#q", tok, text)
		}); err == nil || err == io.EOF {
			t.Errorf("Scan %q: wanted error, but got %v", test, err)
		} else {
			t.Logf("Scan %q: got expected error: %v", test, err)
		}
	}
}
