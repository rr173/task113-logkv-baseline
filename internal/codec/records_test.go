package codec

import "testing"

func TestEncodeDecodeRecords(t *testing.T) {
	data, err := EncodeRecords([]Record{{Key: "a", Value: []byte("one")}, {Key: "b", Deleted: true}})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := DecodeRecords(data)
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	if rows[0].Key != "a" || string(rows[0].Value) != "one" || !rows[1].Deleted {
		t.Fatalf("bad round trip: %#v", rows)
	}
}
