// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

package wirepb

import (
	"io"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestDecoding(t *testing.T) {
	//             @1  #8  ....... @2  ........@3  .   .   .   .   .   @4  ....
	const input = "\012\010abcdefgh\021abcdefgh\030\xc4\x86\x89\x8a\x04\045****"

	want := []struct {
		key  int
		wire wireType
		data string
	}{
		{1, TDelimited, "abcdefgh"},
		{2, TFixed64, "abcdefgh"},
		{3, TVarint, "ABCD"},
		{4, TFixed32, "****"},
	}
	dec := NewDecoder(strings.NewReader(input))
	for i, test := range want {
		t.Logf("Record %d :: %+v", i, test)
		got, err := dec.Next()
		if err != nil {
			t.Fatalf("dec.Next(): unexpected error: %v", err)
		}
		want := &Field{test.key, test.wire, []byte(test.data)}
		if diff := pretty.Compare(got, want); diff != "" {
			t.Errorf("Record %d result differs from expected (-got, +want)\n%s", i, diff)
		}
	}
}

func decode1(s string) *Field {
	f, err := NewDecoder(strings.NewReader(s)).Next()
	if err != nil {
		panic(err)
	}
	return f
}

func TestPacking(t *testing.T) {
	tests := []struct {
		id         int
		wire       wireType
		data, want string
	}{
		{1, TFixed32, "abcd", "\015abcd"},
		{2, TFixed64, "abcdefgh", "\021abcdefgh"},
		{3, TDelimited, "apple pie and cake", "\032\022apple pie and cake"},
		{4, TVarint, "ABCD", " \xc4\x86\x89\x8a\x04"},
		{47, TFixed32, "0123", "\xfd\x020123"},
	}
	for _, test := range tests {
		input := &Field{ID: test.id, Wire: test.wire, Data: []byte(test.data)}
		got := string(input.Pack(nil))
		n := input.Size()

		if len(got) != n {
			t.Errorf("Pack %+v: got length %d, want %d", input, len(got), n)
		}
		if got != test.want {
			t.Errorf("Pack %+v: got %#v, want %#v", input, []byte(got), []byte(test.want))
		}

		rt := decode1(got)
		if diff := pretty.Compare(rt, input); diff != "" {
			t.Errorf("Pack result did not round-trip (-got, +want)\n%s", diff)
		}
	}
}

func TestErrors(t *testing.T) {
	badInputs := []string{
		"\010",       // missing varint length
		"\050\x83",   // malformed varint length
		"\023",       // unsupported wire type
		"\034",       // unsupported wire type
		"\046",       // unknown wire type
		"\052\x0312", // truncated delimited field
		"\061abcdef", // truncated fixed64
		"\075abc",    // truncated fixed32
	}
nextTest:
	for _, input := range badInputs {
		dec := NewDecoder(strings.NewReader(input))
		for {
			f, err := dec.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Logf("Input %q: got expected error: %v", input, err)
				continue nextTest
			}
			t.Logf("Field id=%d, wireType=%d, data=%+v", f.ID, f.Wire, f.Data)
		}
		t.Errorf("Input %q: expected error, but succeeded", input)
	}
}
