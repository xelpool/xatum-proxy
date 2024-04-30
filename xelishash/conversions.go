package xelishash

import (
	"unsafe"
)

// WARNING: this only works on Little Endian architectures
func intInput(input [BYTES_ARRAY_INPUT]byte) *[KECCAK_WORDS]uint64 {
	return (*[KECCAK_WORDS]uint64)(unsafe.Pointer(&input))
}

func scratchpadToSmallpad(s *ScratchPad) *[MEMORY_SIZE * 2]uint32 {
	return (*[MEMORY_SIZE * 2]uint32)(unsafe.Pointer(s))

}

func aesConv(d [16]byte) [4]uint32 {
	return *(*[4]uint32)(unsafe.Pointer(&d))
}
func aesConv2(d [4]uint32) [16]byte {
	return *(*[16]byte)(unsafe.Pointer(&d))
}
