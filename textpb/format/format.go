// Copyright (C) 2015 Michael J. Fromberger. All Rights Reserved.

// Package format renders textpb.Message values to protobuf text format.
package format

import (
	"fmt"
	"io"
	"strings"

	"bitbucket.org/creachadair/pson/textpb"
)

var (
	// Curly
	left  = map[bool]string{false: "<", true: "{"}
	right = map[bool]string{false: ">", true: "}"}
	// Compact
	space = map[bool]string{false: " ", true: ""}
	next  = map[bool]string{false: "\n", true: " "}
	end   = map[bool]string{false: "\n", true: ""}
)

// A Config captures settings for rendering messages in text format.
type Config struct {
	Compact bool   // If true, omit vertical whitespace.
	Curly   bool   // If true, use {} for grouping rather than <>.
	Indent  string // Use this string for each level of indentation.
}

// Text renders the specified message to w in text format.
func (c Config) Text(w io.Writer, msg textpb.Message) error { return c.textMessage(w, msg, 0) }

func (c Config) textMessage(w io.Writer, msg textpb.Message, level int) error {
	for i, field := range msg {
		if err := c.textField(w, field, level, i < len(msg)-1); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) textField(w io.Writer, field *textpb.Field, level int, sep bool) error {
	if len(field.Values) == 0 { // empty: treat as empty repeated field
		return fp(w, c.indent(level), field.Name, c.space(), c.left(), c.right())
	}
	for i, value := range field.Values {
		if err := c.textValue(w, field.Name, value, level, i < len(field.Values)-1); err != nil {
			return err
		}
	}
	return c.next(w, sep)
}

func (c Config) textValue(w io.Writer, name string, value *textpb.Value, level int, sep bool) error {
	if err := fp(w, c.indent(level), name); err != nil {
		return err
	}
	if value.Msg == nil {
		if err := fp(w, ":", c.space(), tokenText(value)); err != nil {
			return err
		}
		return c.next(w, sep)
	}
	if len(value.Msg) == 0 {
		return fp(w, " ", c.left(), c.right())
	} else if err := fp(w, " ", c.left(), c.first()); err != nil {
		return err
	} else if err := c.textMessage(w, value.Msg, level+1); err != nil {
		return err
	} else if err := fp(w, c.last(), c.indent(level), c.right()); err != nil {
		return err
	}
	return c.next(w, sep)
}

func (c Config) indent(level int) string {
	if c.Compact {
		return ""
	} else if c.Indent == "" {
		return strings.Repeat("  ", level)
	}
	return strings.Repeat(c.Indent, level)
}

func tokenText(v *textpb.Value) string {
	if v.Type == textpb.String {
		return `"` + strings.Replace(v.Text, `"`, `\"`, -1) + `"`
	}
	return v.Text
}

func (c Config) left() string  { return left[c.Curly] }
func (c Config) right() string { return right[c.Curly] }
func (c Config) space() string { return space[c.Compact] }
func (c Config) first() string { return end[c.Compact] }
func (c Config) last() string  { return end[c.Compact] }

func (c Config) next(w io.Writer, want bool) error {
	if want {
		return fp(w, next[c.Compact])
	}
	return nil
}

func fp(w io.Writer, args ...interface{}) error {
	_, err := fmt.Fprint(w, args...)
	return err
}
