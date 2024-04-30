package xelisutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
	"xatum-proxy/xelishash"

	"github.com/duggavo/serializer"
)

// Xatum Protocol BlockMiner implementation

const BLOCKMINER_LENGTH = 112

type BlockMiner [BLOCKMINER_LENGTH]byte

func NewBlockMiner(workhash, extranonce, publickey [32]byte) BlockMiner {
	s := serializer.Serializer{
		Endian: binary.BigEndian,
	}

	s.AddFixedByteArray(workhash[:], 32)
	s.AddUint64(uint64(time.Now().UnixMilli()))
	s.AddUint64(0)
	s.AddFixedByteArray(extranonce[:], 32)
	s.AddFixedByteArray(publickey[:], 32)

	return BlockMiner(s.Data)

}
func NewBlockMinerFromBlob(blob []byte) (BlockMiner, error) {
	if len(blob) != 96 {
		return BlockMiner{}, errors.New("malformed BlockMinerBlob")
	}

	return NewBlockMiner([32]byte(blob[0:32]), [32]byte(blob[32:32*2]), [32]byte(blob[32*2:32*3])), nil

}

// SETTER methods
func (b *BlockMiner) SetTimestamp(t uint64) {
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, t)

	// update the timestamp
	b[32] = tb[0]
	b[33] = tb[1]
	b[34] = tb[2]
	b[35] = tb[3]
	b[36] = tb[4]
	b[37] = tb[5]
	b[38] = tb[6]
	b[39] = tb[7]
}
func (b *BlockMiner) SetNonce(n uint64) {
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, n)

	// update the nonce
	b[40] = tb[0]
	b[41] = tb[1]
	b[42] = tb[2]
	b[43] = tb[3]
	b[44] = tb[4]
	b[45] = tb[5]
	b[46] = tb[6]
	b[47] = tb[7]
}

func (b *BlockMiner) SetExtraNonce(n [32]byte) {
	for i := 0; i < 32; i++ {
		b[48+i] = n[i]
	}
}

// returns true if the first 28 bytes of the extra nonces match
func ValidateExtraNonces(a, b [32]byte) bool {
	return bytes.Equal(a[:32-4], b[:32-4])
}

// GETTER methods

func (b BlockMiner) Serialize() []byte {
	return b[:]
}
func (b BlockMiner) Hash() [32]byte {
	return FastHash(b[:])
}
func (b BlockMiner) PowHash(sp *xelishash.ScratchPad) [32]byte {
	return PowHash(b[:], sp)
}

func (b BlockMiner) GetWorkhash() [32]byte {
	return [32]byte(b[:32])
}
func (b BlockMiner) GetTimestamp() uint64 {
	return binary.BigEndian.Uint64(b[32:40])
}
func (b BlockMiner) GetNonce() uint64 {
	return binary.BigEndian.Uint64(b[40:48])
}
func (b BlockMiner) GetExtraNonce() [32]byte {
	return [32]byte(b[48:80])
}
func (b BlockMiner) GetPublickey() [32]byte {
	return [32]byte(b[80:112])
}

func (b BlockMiner) GetBlob() []byte {
	wh := b.GetWorkhash()
	xn := b.GetExtraNonce()
	pk := b.GetPublickey()
	return append(append(wh[:], xn[:]...), pk[:]...)
}
