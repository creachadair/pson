// Package wirepb supports decoding raw wire-format protobuf messages, where
// "raw" means the decoding is done without knowledge of the schema.
package wirepb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// A Decoder consumes input from an io.Reader pointing to a wire-format
// protobuf message.
type Decoder struct {
	buf *bufio.Reader
}

// NewDecoder creates a new decoder that reads data from r.
func NewDecoder(r io.Reader) Decoder { return Decoder{bufio.NewReader(r)} }

// A Field represents a field read from a wire-format message.  The data in the
// field are returned as encoded. Further decoding into a higher-level schema
// is the caller's responsibility.
type Field struct {
	ID   int
	Wire Type
	Data []byte
}

// Next returns the next field in the message.
func (d Decoder) Next() (*Field, error) {
	v, err := binary.ReadUvarint(d.buf)
	if err != nil {
		return nil, err
	}
	f := &Field{
		ID:   int(v >> 3),
		Wire: Type(v & 7),
	}

	switch f.Wire {
	case TVarint:
		w, err := binary.ReadUvarint(d.buf)
		if err != nil {
			return nil, err
		}
		var buf [8]byte
		i := 0
		for ; w > 0; w >>= 8 {
			buf[i] = byte(w & 255)
			i++
		}
		f.Data = buf[:i]
		return f, nil

	case TFixed64:
		f.Data = make([]byte, 8)

	case TDelimited:
		w, err := binary.ReadUvarint(d.buf)
		if err != nil {
			return nil, err
		}
		f.Data = make([]byte, w)

	case TFixed32:
		f.Data = make([]byte, 4)

	default:
		return nil, fmt.Errorf("unknown wire type %d", f.Wire)
	}
	if _, err := io.ReadFull(d.buf, f.Data); err != nil {
		return nil, err
	}
	return f, nil
}

// A Type represents the wire type of a field key
type Type int

const (
	TVarint     Type = 0 // varint-encoded value
	TFixed64    Type = 1 // fixed-width 64-bit value (LSB first)
	TDelimited  Type = 2 // length-prefixed value (varint + bytes)
	TStartGroup Type = 3 // deprecated, unused
	TEndGroup   Type = 4 // deprecated, unused
	TFixed32    Type = 5 // fixed-width 32-bit value (LSB first)

	lastType = TFixed32
)
