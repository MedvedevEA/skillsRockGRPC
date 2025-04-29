package secure

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256
func GetHash(text string) string {
	hash := sha256.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}
func CheckHash(text string, hash string) bool {
	textHash := GetHash(text)
	return textHash == hash
}
