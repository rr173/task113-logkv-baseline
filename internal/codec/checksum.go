package codec

import (
	"encoding/binary"
	"hash/crc32"
)

func Checksum(data []byte) uint32 { return crc32.ChecksumIEEE(data) }
func WithChecksum(data []byte) []byte {
	result := make([]byte, len(data)+4)
	copy(result, data)
	binary.BigEndian.PutUint32(result[len(data):], Checksum(data))
	return result
}
func VerifyChecksum(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	body := data[:len(data)-4]
	return binary.BigEndian.Uint32(data[len(body):]) == Checksum(body)
}
