// Package codec implements the on-disk / wire record format used by backup,
// restore and export operations of the logkv store.
package codec

import (
	"encoding/binary"
	"errors"
	"io"
)

// Record is a single serialized key/value entry.
type Record struct {
	Key       string
	Value     []byte
	ExpiresAt int64
	Deleted   bool
}

// ErrCorruptStream is returned when the decoder reads a malformed record.
var ErrCorruptStream = errors.New("codec: corrupt stream")

// Encoder writes records to an underlying writer in a length-prefixed binary
// format. It is not safe for concurrent use.
type Encoder struct {
	w   io.Writer
	buf []byte
}

// NewEncoder constructs an Encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, buf: make([]byte, 8)}
}

func (e *Encoder) writeUint64(v uint64) error {
	binary.BigEndian.PutUint64(e.buf, v)
	_, err := e.w.Write(e.buf)
	return err
}

func (e *Encoder) writeBytes(b []byte) error {
	if err := e.writeUint64(uint64(len(b))); err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	_, err := e.w.Write(b)
	return err
}

// Write serializes a single record.
func (e *Encoder) Write(r Record) error {
	var deleted byte
	if r.Deleted {
		deleted = 1
	}
	if err := e.writeBytes([]byte(r.Key)); err != nil {
		return err
	}
	if err := e.writeBytes(r.Value); err != nil {
		return err
	}
	if err := e.writeUint64(uint64(r.ExpiresAt)); err != nil {
		return err
	}
	if _, err := e.w.Write([]byte{deleted}); err != nil {
		return err
	}
	return nil
}

// Decoder reads records produced by Encoder. It is not safe for concurrent use.
type Decoder struct {
	r   io.Reader
	buf []byte
}

// NewDecoder constructs a Decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, buf: make([]byte, 8)}
}

func (d *Decoder) readUint64() (uint64, error) {
	if _, err := io.ReadFull(d.r, d.buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(d.buf), nil
}

func (d *Decoder) readBytes() ([]byte, error) {
	n, err := d.readUint64()
	if err != nil {
		return nil, err
	}
	if n > 1<<30 {
		return nil, ErrCorruptStream
	}
	if n == 0 {
		return []byte{}, nil
	}
	b := make([]byte, int(n))
	if _, err := io.ReadFull(d.r, b); err != nil {
		return nil, err
	}
	return b, nil
}

// Read deserializes the next record. It returns io.EOF when the stream ends.
func (d *Decoder) Read() (Record, error) {
	key, err := d.readBytes()
	if err != nil {
		return Record{}, err
	}
	value, err := d.readBytes()
	if err != nil {
		return Record{}, err
	}
	expiresAt, err := d.readUint64()
	if err != nil {
		return Record{}, err
	}
	flag := make([]byte, 1)
	if _, err := io.ReadFull(d.r, flag); err != nil {
		return Record{}, err
	}
	return Record{
		Key:       string(key),
		Value:     value,
		ExpiresAt: int64(expiresAt),
		Deleted:   flag[0] == 1,
	}, nil
}
