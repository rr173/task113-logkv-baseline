package codec

import (
	"bytes"
	"fmt"
	"io"
)

func ValidateRecord(r Record) error {
	if r.Key == "" {
		return fmt.Errorf("codec: empty key")
	}
	if r.ExpiresAt < 0 {
		return fmt.Errorf("codec: negative expiry")
	}
	return nil
}

func EncodeRecords(records []Record) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	for _, record := range records {
		if err := ValidateRecord(record); err != nil {
			return nil, err
		}
		if err := enc.Write(record); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func DecodeRecords(data []byte) ([]Record, error) {
	dec := NewDecoder(bytes.NewReader(data))
	out := make([]Record, 0)
	for {
		record, err := dec.Read()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
}
