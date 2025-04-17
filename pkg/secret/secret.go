package secret

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func GenerateKey(keySize int) (*rsa.PrivateKey, ssh.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return priv, pub, nil
}

func MarshalRSAPrivate(priv *rsa.PrivateKey, typePrivateKey string) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type: typePrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}
func MarshalRSAPublic(pub ssh.PublicKey) []byte {
	return bytes.TrimSuffix(ssh.MarshalAuthorizedKey(pub), []byte{'\n'})
}

func UnmarshalRSAPrivate(bytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
func UnmarshalRSAPublic(bytes []byte) (ssh.PublicKey, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey(bytes)
	return pub, err
}
