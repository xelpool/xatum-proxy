package xelishash

import (
	"math/bits"
)

const MEMORY_SIZE = 32768
const SCRATCHPAD_ITERS = 5000
const ITERS = 1
const BUFFER_SIZE = 42
const SLOT_LENGTH = 256

// Untweakable parameters
const KECCAK_WORDS = 25
const BYTES_ARRAY_INPUT = KECCAK_WORDS * 8
const HASH_SIZE = 32

const STAGE_1_MAX = MEMORY_SIZE / KECCAK_WORDS

type ScratchPad [MEMORY_SIZE]uint64
type Hash [HASH_SIZE]byte

func stage_1(int_input *[KECCAK_WORDS]uint64, scratch_pad *ScratchPad, a0 uint64, a1 uint64, b0 uint64, b1 uint64) {
	for i := a0; i <= a1; i++ {
		keccakp(int_input)

		var rand_int uint64 = 0
		for j := b0; j <= b1; j++ {
			pair_idx := (j + 1) % KECCAK_WORDS
			pair_idx2 := (j + 2) % KECCAK_WORDS

			target_idx := i*KECCAK_WORDS + j
			a := int_input[j] ^ rand_int
			// Branching
			left := int_input[pair_idx]
			right := int_input[pair_idx2]
			xor := left ^ right

			var v uint64
			switch xor & 0x3 {
			case 0:
				v = left & right
			case 1:
				v = ^(left & right)
			case 2:
				v = ^xor
			case 3:
				v = xor
			}

			b := a ^ v
			rand_int = b
			scratch_pad[target_idx] = b
		}
	}
}

func XelisHash(input []byte, scratch_pad *ScratchPad) ([32]byte, error) {
	var int_input *[KECCAK_WORDS]uint64 = intInput([BYTES_ARRAY_INPUT]byte(input[:BYTES_ARRAY_INPUT]))

	// stage 1
	stage_1(int_input, scratch_pad, 0, STAGE_1_MAX-1, 0, KECCAK_WORDS-1)
	stage_1(int_input, scratch_pad, STAGE_1_MAX, STAGE_1_MAX, 0, 17)

	// stage 2

	// this is equal to MEMORY_SIZE, just in u32 format
	var slots [SLOT_LENGTH]uint32

	var small_pad *[MEMORY_SIZE * 2]uint32 = scratchpadToSmallpad(scratch_pad)

	slots = [SLOT_LENGTH]uint32(small_pad[len(small_pad)-SLOT_LENGTH:])

	var indices [SLOT_LENGTH]uint16

	for i := 0; i < ITERS; i++ {
		for j := 0; j < len(small_pad)/SLOT_LENGTH; j++ {
			// Initialize indices
			for k := 0; k < SLOT_LENGTH; k++ {
				indices[k] = uint16(k)
			}

			for slot_idx := SLOT_LENGTH - 1; slot_idx >= 0; slot_idx-- {

				index_in_indices := int((small_pad[j*SLOT_LENGTH+slot_idx] % (uint32(slot_idx) + 1)))
				index := int(indices[index_in_indices])
				indices[index_in_indices] = indices[slot_idx]

				// THIS IS THE MOST PERFORMANCE-CRITICAL SECTION

				// Split the loop in two to avoid checking k == index
				sum := slots[index]
				offset := j * SLOT_LENGTH
				for k := 0; k < SLOT_LENGTH; k++ {
					if k == index {
						continue
					}

					if slots[k]>>31 == 0 {
						sum = sum + small_pad[offset+k]
					} else {
						sum = sum - small_pad[offset+k]
					}
				}

				slots[index] = sum
			}
		}
	}

	copy(small_pad[(MEMORY_SIZE*8/4)-SLOT_LENGTH:], slots[:])

	// stage 3
	var key [16]byte
	var block [16]byte

	addr_a := (scratch_pad[MEMORY_SIZE-1] >> 15) & 0x7FFF
	addr_b := scratch_pad[MEMORY_SIZE-1] & 0x7FFF

	var mem_buffer_a [BUFFER_SIZE]uint64
	var mem_buffer_b [BUFFER_SIZE]uint64

	for i := 0; i < BUFFER_SIZE; i++ {
		mem_buffer_a[i] = scratch_pad[((addr_a + uint64(i)) % MEMORY_SIZE)]
		mem_buffer_b[i] = scratch_pad[((addr_b + uint64(i)) % MEMORY_SIZE)]
	}

	var final_result Hash

	for i := 0; i < SCRATCHPAD_ITERS; i++ {
		mem_a := mem_buffer_a[i%BUFFER_SIZE]
		mem_b := mem_buffer_b[i%BUFFER_SIZE]

		copy(block[:8], toLE(mem_b))
		copy(block[8:], toLE(mem_a))

		block = aesRound2(block, key)

		hash1 := fromLE(block[0:8])
		hash2 := mem_a ^ mem_b

		result := ^(hash1 ^ hash2)

		for j := 0; j < HASH_SIZE; j++ {
			a := mem_buffer_a[(j+i)%BUFFER_SIZE]
			b := mem_buffer_b[(j+i)%BUFFER_SIZE]

			// more branching
			switch (result >> (j * 2)) & 0xf {
			case 0:
				result = bits.RotateLeft64(result, j) ^ b
			case 1:
				result = ^(bits.RotateLeft64(result, j) ^ a)
			case 2:
				result = ^(result ^ a)
			case 3:
				result = result ^ b
			case 4:
				result = result ^ (a + b)
			case 5:
				result = result ^ (a - b)
			case 6:
				result = result ^ (b - a)
			case 7:
				result = result ^ (a * b)
			case 8:
				result = result ^ (a & b)
			case 9:
				result = result ^ (a | b)
			case 10:
				result = result ^ (a ^ b)
			case 11:
				result = result ^ (a - result)
			case 12:
				result = result ^ (b - result)
			case 13:
				result = result ^ (a + result)
			case 14:
				result = result ^ (result - a)
			case 15:
				result = result ^ (result - b)
			}
		}

		addr_b = result & 0x7FFF
		mem_buffer_a[i%BUFFER_SIZE] = result
		mem_buffer_b[i%BUFFER_SIZE] = scratch_pad[addr_b]

		addr_a = (result >> 15) & 0x7FFF
		scratch_pad[addr_a] = result

		index := SCRATCHPAD_ITERS - i - 1
		if index < 4 {
			copy(final_result[index*8:(SCRATCHPAD_ITERS-i)*8], toBE(result))
		}
	}

	return final_result, nil
}
