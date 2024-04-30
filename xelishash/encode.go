package xelishash

import "encoding/binary"

func toLE(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func fromLE(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
func toBE(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}
