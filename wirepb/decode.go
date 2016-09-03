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

// Next returns the next field in the message.
func (d Decoder) Next() (*Field, error) {
	v, err := binary.ReadUvarint(d.buf)
	if err != nil {
		return nil, err
	}
	f := &Field{
		ID:   int(v >> 3),
		Wire: wireType(v & 7),
	}

	switch f.Wire {
	case TVarint:
		w, err := binary.ReadUvarint(d.buf)
		if err != nil {
			return nil, err
		}
		f.Data = Uint64(w)
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

// A wireType represents the wire type of a field key
type wireType int

const (
	TVarint     wireType = 0 // varint-encoded value
	TFixed64    wireType = 1 // fixed-width 64-bit value (LSB first)
	TDelimited  wireType = 2 // length-prefixed value (varint + bytes)
	TStartGroup wireType = 3 // deprecated, unused
	TEndGroup   wireType = 4 // deprecated, unused
	TFixed32    wireType = 5 // fixed-width 32-bit value (LSB first)
)

// A Field represents a field read from a wire-format message.  The data in the
// field are returned as encoded. Further decoding into a higher-level schema
// is the caller's responsibility.
type Field struct {
	ID   int
	Wire wireType
	Data []byte
}

// Size reports the number of bytes needed to encode f in wire format, or 0 if
// f cannot be encoded.
func (f *Field) Size() int {
	n := varintSize(uint64(f.ID) << 3)
	switch f.Wire {
	case TVarint:
		return n + (8*len(f.Data)+6)/7
	case TFixed64:
		return n + 8
	case TDelimited:
		return n + varintSize(uint64(len(f.Data))) + len(f.Data)
	case TFixed32:
		return n + 4
	default:
		return 0
	}
}

// Pack encodes f in wire format and appends the result to buf, allowing the
// caller to control allocation.
func (f *Field) Pack(buf []byte) []byte {
	var bits [10]byte // buffer for varint encoding

	// Pack the field ID.
	key := (uint64(f.ID) << 3) | uint64(f.Wire)
	n := binary.PutUvarint(bits[:], key)

	// Pack the field value.
	return f.PackValue(append(buf, bits[:n]...))
}

// PackValue encodes the value of f in wire format and appends the result to
// buf, allowing the caller to control allocation. Returns nil if f cannot be
// packed.
func (f *Field) PackValue(buf []byte) []byte {
	var bits [10]byte // buffer for varint encoding

	// Pack the field value
	switch f.Wire {
	case TVarint:
		return append(buf, dataToVarint(f.Data)...)

	case TDelimited:
		n := binary.PutUvarint(bits[:], uint64(len(f.Data)))
		buf = append(buf, bits[:n]...)
		return append(buf, f.Data...)

	case TFixed64:
		return appendN(buf, f.Data, 8)

	case TFixed32:
		return appendN(buf, f.Data, 4)

	default:
		return nil
	}
}

// appendN appends up to n bytes from data to old, padding with zeroes if
// len(data) < n, and returns the expanded slice.
func appendN(old, data []byte, n int) []byte {
	t := len(data)
	if t > n {
		t = n
	}
	old = append(old, data[:t]...)
	for t < n {
		old = append(old, 0)
		t++
	}
	return old
}

// varintSize reports the number of bytes needed to encode v as a varint.
func varintSize(v uint64) int {
	n := 1
	for v >>= 7; v != 0; v >>= 7 {
		n++
	}
	return n
}

// Uint64 packs v into a slice of bytes in big-endian order, without any
// unnecessary leading zeroes.
func Uint64(v uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	for i, b := range buf {
		if b != 0 {
			return buf[i:]
		}
	}
	return buf[:1]
}

func dataToVarint(data []byte) []byte {
	var w uint64
	for _, b := range data {
		w = (w << 8) | uint64(b)
	}
	var bits [10]byte
	n := binary.PutUvarint(bits[:], w)
	return bits[:n]
}
