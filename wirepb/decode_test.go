package wirepb

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestDecoding(t *testing.T) {
	//             @1  #8  ....... @2  ........@3  .   .   .   .   .   .   .   .   .   @4  ....
	const input = "\012\010abcdefgh\021abcdefgh\030\xe1\xc4\x8d\xa3\xd6\xcc\xd9\xb3\x68\045ABCD"

	want := []struct {
		key  int
		wire Type
		data string
	}{
		{1, TDelimited, "abcdefgh"},
		{2, TFixed64, "abcdefgh"},
		{3, TVarint, "abcdefgh"},
		{4, TFixed32, "ABCD"},
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
