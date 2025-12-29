package userhasher

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sha256 returns a SHA-256 hash of the given number.
// (previously was email, and number — decided against it, see ADR.MD)
func Sha256(number string) string {
	hash := sha256.New()
	hash.Write([]byte(number))
	return hex.EncodeToString(hash.Sum(nil))
}
