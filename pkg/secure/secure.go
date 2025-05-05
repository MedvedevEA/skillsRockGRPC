package secure

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
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
func LoadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	privateKeyByteArray, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	privateKeyPemBlock, _ := pem.Decode(privateKeyByteArray)
	if privateKeyPemBlock == nil || privateKeyPemBlock.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("decoding error. The PEM block was not found or the type is not equal to RSA PRIVATE KEY")
	}
	return x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
}
