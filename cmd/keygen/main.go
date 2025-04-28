package main

import (
	"flag"
	"log"
	"os"
	"skillsRockGRPC/pkg/secure"
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

	privateKey, publicKey, err := secure.GenerateKey(keySize)
	if err != nil {
		log.Fatal(err)
	}
	privateFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer privateFile.Close()
	_, err = privateFile.Write(secure.MarshalRSAPrivate(privateKey, typePrivateKey))
	if err != nil {
		log.Fatal(err)
	}

	publicFile, err := os.Create(filename + ".pub")
	if err != nil {
		log.Fatal(err)
	}
	_, err = publicFile.Write(secure.MarshalRSAPublic(publicKey))
	if err != nil {
		log.Fatal(err)
	}

}
