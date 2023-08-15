package celestia

import (
	"encoding/hex"
	"math/rand"
	"time"
)

func GenerateRandNamespaceID() string {
	rand.Seed(time.Now().UnixNano())
	nID := make([]byte, 10)
	_, err := rand.Read(nID)
	if err != nil {
		panic(err)
	}
	return "0x" + hex.EncodeToString(nID)
}
