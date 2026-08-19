package codec

import (
	"io"
)

type StreamStats struct {
	Records int
	Bytes   int
}

func CopyRecords(dst io.Writer, src io.Reader) (StreamStats, error) {
	dec := NewDecoder(src)
	enc := NewEncoder(dst)
	out := StreamStats{}
	for {
		record, err := dec.Read()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		if err := ValidateRecord(record); err != nil {
			return out, err
		}
		if err := enc.Write(record); err != nil {
			return out, err
		}
		out.Records++
		out.Bytes += len(record.Key) + len(record.Value)
	}
}
