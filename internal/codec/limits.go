package codec

const MaxRecordBytes = 1 << 30

func RecordBytes(r Record) int { return len(r.Key) + len(r.Value) + 17 }
func FitsRecord(r Record) bool { return RecordBytes(r) <= MaxRecordBytes }
func CountBytes(records []Record) int {
	total := 0
	for _, record := range records {
		total += RecordBytes(record)
	}
	return total
}
