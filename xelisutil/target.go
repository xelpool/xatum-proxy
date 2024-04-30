package xelisutil

import (
	"bytes"
	"math/big"
	"xatum-proxy/util"
)

var maxBigInt *big.Int

func init() {

	b, _ := big.NewInt(0).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	maxBigInt = b

}

func GetTarget(diff uint64) *big.Int {
	if diff == 0 {
		return big.NewInt(0)
	}

	diffBigInt := big.NewInt(0)
	diffBigInt = diffBigInt.SetBytes(util.Uint64ToBigEndian(diff))

	return diffBigInt.Div(maxBigInt, diffBigInt)
}

func GetTargetBytes(diff uint64) [32]byte {
	data := make([]byte, 32)

	byt := GetTarget(diff).Bytes()

	copy(data[32-len(byt):], byt)

	return [32]byte(data)
}

// returns true if the hash matches difficulty
func CheckDiff(hash [32]byte, diff uint64) bool {
	target := GetTargetBytes(diff)

	return bytes.Compare(hash[:], target[:]) < 0
}
