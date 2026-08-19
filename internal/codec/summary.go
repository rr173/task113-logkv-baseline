package codec

type Summary struct {
	Records  int
	Deleted  int
	Bytes    int
	Expiring int
}

func Summarize(records []Record) Summary {
	out := Summary{Records: len(records)}
	for _, record := range records {
		out.Bytes += RecordBytes(record)
		if record.Deleted {
			out.Deleted++
		}
		if record.ExpiresAt > 0 {
			out.Expiring++
		}
	}
	return out
}
