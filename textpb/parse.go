// Package textpb implements a scanner and parser for text-format protobuf
// messages.
package textpb

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// A Message represents a collection of named fields.
type Message []*Field

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
	var out Message
	for _, field := range names {
		out = append(out, field)
	}
	sort.Sort(out)
	return out
}

func (m Message) Len() int           { return len(m) }
func (m Message) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Message) Less(i, j int) bool { return m[i].Name < m[j].Name }

// A Field represents a named field. The value of a field may be a message or a
// collection of values.
type Field struct {
	Name   string
	Values []*Value
}

func (f *Field) String() string { return fmt.Sprintf("#<field name=%q values=%+v>", f.Name, f.Values) }

// A Value represents the value of a field. If Msg is non-nil, the other
// fields will be ignored.
type Value struct {
	Msg  Message
	Type Token
	Text string
}

func (v *Value) combine() *Value {
	if v.Msg == nil {
		return v
	}
	return &Value{Msg: v.Msg.Combine()}
}

func (v *Value) String() string {
	if v.Msg != nil {
		return fmt.Sprintf("#<message %+v>", v.Msg)
	}
	return fmt.Sprintf("#<value %v %q>", v.Type, v.Text)
}

// Int returns the value as an integer, if possible.
func (v *Value) Int() (int, error) { return strconv.Atoi(v.Text) }

// Number returns the value as a floating-point number, if possible.
func (v *Value) Number() (float64, error) { return strconv.ParseFloat(v.Text, 64) }

// Bool returns the value as a Boolean, if possible.
func (v *Value) Bool() (bool, error) {
	switch v.Text {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("invalid bool value")
	}
}

// Parse parses the input from r and returns a Message that represents it.
func Parse(r io.Reader) (Message, error) {
	p := parser{NewScanner(r)}
	if !p.Next() {
		if err := p.Err(); err != nil && err != io.EOF {
			return nil, p.fail(err.Error())
		}
		return nil, nil
	}
	return p.parseMessage(None)
}

type parser struct {
	*Scanner
}

func (p parser) fail(msg string, args ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf("line %d: ", p.Line())+msg, args...)
}

func (p parser) parseMessage(until Token) (Message, error) {
	var msg Message
	for {
		tok := p.Token()
		if tok == until {
			return msg, nil
		} else if tok != Name && tok != TypeName {
			return nil, p.fail("found %v, wanted name or type", tok)
		}
		name := p.Text()

		if !p.Next() {
			return nil, p.fail("found %v, wanted %v or message", tok, Colon)
		}
		var field *Field
		var err error
		switch p.Token() {
		case LeftA:
			field, err = p.parseMessageField(name, RightA)
		case LeftC:
			field, err = p.parseMessageField(name, RightC)
		case Colon:
			field, err = p.parseValueOrMessage(name)
		default:
			return nil, p.fail("found %v, wanted %v or message", p.Token(), Colon)
		}
		if err != nil {
			return nil, err
		}
		if tok == TypeName && field.Values[0].Msg == nil {
			return nil, p.fail("type name %q requires a message value", name)
		}
		msg = append(msg, field)
		if tok := p.Token(); tok == Comma || tok == Semi {
			p.Next() // skip optional separator
		}
	}
	if err := p.Err(); err == io.EOF {
		if until != None {
			return nil, p.fail("wanted %v, but found end of input", until)
		}
	} else if err != nil {
		return nil, err
	}
	return msg, nil
}

func (p parser) parseMessageField(name string, until Token) (*Field, error) {
	if !p.Next() {
		return nil, p.fail("%v: wanted field or %v", p.Err(), until)
	}
	msg, err := p.parseMessage(until)
	if err != nil {
		return nil, err
	}
	if tok := p.Token(); tok != until {
		return nil, p.fail("found %v, wanted %v", tok, until)
	}
	p.Next()
	return &Field{
		Name:   name,
		Values: []*Value{{Msg: msg}},
	}, nil
}

func (p parser) parseValueOrMessage(name string) (*Field, error) {
	if !p.Next() {
		return nil, p.fail("%v: wanted value or message for %q", p.Err(), name)
	}
	tok := p.Token()
	if tok == LeftA {
		return p.parseMessageField(name, RightA)
	} else if tok == LeftC {
		return p.parseMessageField(name, RightC)
	} else if !tok.IsValue() {
		return nil, p.fail("unexpected %v, wanted a value", tok)
	}
	out := &Field{
		Name: name,
		Values: []*Value{{
			Type: tok,
			Text: p.Text(),
		}},
	}

	// Consecutive string literal tokens are concatenated.
	for p.Next() {
		if p.Token() == String && tok == String {
			out.Values[0].Text += p.Text()
			continue
		}
		break
	}
	return out, nil
}
