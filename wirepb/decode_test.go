package wirepb

import (
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
