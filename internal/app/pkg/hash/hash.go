package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

func GenSHA1(in string) string {
	h := sha1.New()
	h.Write([]byte(in))
	return hex.EncodeToString(h.Sum(nil))
}
