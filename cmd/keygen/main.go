package main

import (
	"flag"
	"log"
	"os"
	"skillsRockGRPC/pkg/secret"
)

// Генерация и сохранение в файл пары ключей
func main() {
	var (
		filename       string
		keySize        int
		typePrivateKey string
	)

	flag.StringVar(&filename, "filename", "rsa", "file name")
	flag.IntVar(&keySize, "keysize", 1024, " key size")
	flag.StringVar(&typePrivateKey, "typeprivatekey", "RSA PRIVATE KEY", "type private key")
	flag.Parse()

	privateKey, publicKey, err := secret.GenerateKey(keySize)
	if err != nil {
		log.Fatal(err)
	}
	privateFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer privateFile.Close()
	_, err = privateFile.Write(secret.MarshalRSAPrivate(privateKey, typePrivateKey))
	if err != nil {
		log.Fatal(err)
	}

	publicFile, err := os.Create(filename + ".pub")
	if err != nil {
		log.Fatal(err)
	}
	_, err = publicFile.Write(secret.MarshalRSAPublic(publicKey))
	if err != nil {
		log.Fatal(err)
	}

}
