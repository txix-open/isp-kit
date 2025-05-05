package requestid

import (
	"crypto/rand"
	"encoding/hex"
)

const requestIdLength = 16

func Next() string {
	value := make([]byte, requestIdLength)
	_, err := rand.Read(value)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
