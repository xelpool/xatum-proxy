package xelishash

import (
	"fmt"
	"testing"
)

func testInput(input []byte, expected_hash [32]byte) error {
	var scratch_pad ScratchPad
	hash, err := XelisHash(input, &scratch_pad)
	if err != nil {
		return err
	}

	if hash != expected_hash {
		return fmt.Errorf("hash %x does not match expected hash %x", hash, expected_hash)
	}

	return nil
}

func TestHash(t *testing.T) {
	t.Log("testing hash")

	var int_input [KECCAK_WORDS]uint64

	keccakp(&int_input)

	t.Log(int_input)

	err := testInput(make([]byte, 200),
		[32]byte{0x0e, 0xbb, 0xbd, 0x8a, 0x31, 0xed, 0xad, 0xfe, 0x09, 0x8f, 0x2d, 0x77, 0x0d, 0x84,
			0xb7, 0x19, 0x58, 0x86, 0x75, 0xab, 0x88, 0xa0, 0xa1, 0x70, 0x67, 0xd0, 0x0a, 0x8f,
			0x36, 0x18, 0x22, 0x65})
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, 200)

	copy(data, []byte("xelis-hashing-algorithm"))

	err = testInput(data, [32]byte{
		106, 106, 173, 8, 207, 59, 118, 108, 176, 196, 9, 124, 250, 195, 3,
		61, 30, 146, 238, 182, 88, 83, 115, 81, 139, 56, 3, 28, 176, 86, 68, 21})
	if err != nil {
		t.Fatal(err)
	}

}

func BenchmarkHash(b *testing.B) {
	var scratch_pad ScratchPad

	var input = make([]byte, 200)

	b.Log(b.N)

	for i := 0; i < b.N*200; i++ {
		XelisHash(input, &scratch_pad)

	}
}
