package textpb

import (
	"bytes"
	"fmt"
	"strconv"
)

// MarshalJSON implements the json.Marshaler interface.  Conversion to JSON is
// entirely lexical; the parser does not know anything about the original
// structure of the protobuf message, it just uses the structure of the text
// format to infer the names for things.
//
// Conventions:
//   A field with no value is represented by "null".
//   Booleans are represented by "true" and "false".
//   Numbers are copied literally.
//   Field names and enumerators are encoded as strings.
//   Repeated fields are encoded as lists.
//
// Note that we don't really know which fields are declared as repeated; we
// assume a field is repeated if it has 0 or > 1 values.
func (m Message) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if err := m.marshalJSON(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m Message) marshalJSON(buf *bytes.Buffer) error {
	buf.WriteByte('{')
	for i, f := range m {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "%q:", f.Name)
		if len(f.Values) != 1 {
			buf.WriteByte('[')
		}
		for j, v := range f.Values {
			if j > 0 {
				buf.WriteByte(',')
			}
			if err := v.marshalJSON(buf); err != nil {
				return err
			}
		}
		if len(f.Values) != 1 {
			buf.WriteByte(']')
		}
	}
	buf.WriteByte('}')
	return nil
}

func (v *Value) marshalJSON(buf *bytes.Buffer) error {
	if v.Msg != nil {
		return v.Msg.marshalJSON(buf)
	}
	switch v.Type {
	case None:
		buf.WriteString("null")
	case Name, String:
		buf.WriteString(strconv.Quote(v.Text))
	case TypeName:
		buf.WriteString(strconv.Quote("[" + v.Text + "]"))
	case True, False, Number:
		buf.WriteString(v.Text)
	default:
		return fmt.Errorf("invalid value type: %v", v.Type)
	}
	return nil
}