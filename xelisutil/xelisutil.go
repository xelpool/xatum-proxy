package xelisutil

import (
	"xatum-proxy/xelishash"

	"github.com/zeebo/blake3"
)

func FastHash(d []byte) [32]byte {
	return blake3.Sum256(d)
}

func PowHash(d []byte, scratchpad *xelishash.ScratchPad) [32]byte {
	if len(d) > xelishash.BYTES_ARRAY_INPUT {
		panic("PowHash input is too long")
	}

	buf := make([]byte, xelishash.BYTES_ARRAY_INPUT)

	copy(buf, d)

	data, err := xelishash.XelisHash(buf, scratchpad)

	if err != nil {
		panic(err)
	}

	return data
}
