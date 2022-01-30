package tpl

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/gob"
)

const hashLength = 16 // md5 hash length

type Hash [hashLength]byte

func HashStruct(s interface{}) (Hash, error) {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(s); err != nil {
		return Hash{}, err
	}

	return md5.Sum(b.Bytes()), nil //nolint:gosec
}

func HashBytes(b []byte) Hash {
	return md5.Sum(b) //nolint:gosec
}
