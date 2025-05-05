// Генерация и сохранение в файл пары ключей
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"os"
)

func main() {
	var (
		privateKeyFileName string
		publicKeyFileName  string
		keySize            int
	)

	flag.StringVar(&privateKeyFileName, "private", "private.pem", "private key file name")
	flag.StringVar(&publicKeyFileName, "public", "public.pem", "public key file name")
	flag.IntVar(&keySize, "keysize", 1024, " key size")
	flag.Parse()
	// Generate RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}
	// Save the private key to a file
	privateKeyFile, err := os.Create(privateKeyFileName)
	if err != nil {
		log.Fatalf("Failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	privateKeyPem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPem); err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}
	// Extract the public key and save to a file
	publicKeyFile, err := os.Create(publicKeyFileName)
	if err != nil {
		log.Fatalf("Failed to create public key file: %v", err)
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}
	publicKeyPem := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := pem.Encode(publicKeyFile, publicKeyPem); err != nil {
		log.Fatalf("Failed to write public key: %v", err)
	}

}
