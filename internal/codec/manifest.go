package codec

import "fmt"

type Manifest struct {
	Version  int
	Records  int
	Bytes    int
	Checksum uint32
}

func BuildManifest(records []Record) (Manifest, error) {
	data, err := EncodeRecords(records)
	if err != nil {
		return Manifest{}, err
	}
	for _, record := range records {
		if !FitsRecord(record) {
			return Manifest{}, fmt.Errorf("record too large: %s", record.Key)
		}
	}
	return Manifest{Version: 1, Records: len(records), Bytes: len(data), Checksum: Checksum(data)}, nil
}
