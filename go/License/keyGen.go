//生成公钥和私钥pem文件
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
	var bits int
	flag.IntVar(&bits, "b", 1024, "key length")
	if err := GenRsaKey(bits, "public.pem", "private.pem"); err != nil {
		log.Fatal("Genarate public.pem & private.pem error:", err.Error())
	}
	log.Println("Genarate public.pem & private.pem OK!")
}

/* 生成 私钥和公钥 */
func GenRsaKey(bits int, publicFile, privateFile string) error {
	//生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}

	privateFd, err := os.Create(privateFile)
	if err != nil {
		return err
	}
	defer privateFd.Close()
	err = pem.Encode(privateFd, block)
	if err != nil {
		return err
	}

	//生成公钥文件
	publicKey := &privateKey.PublicKey
	defPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: defPkix,
	}
	publicFd, err := os.Create(publicFile)
	if err != nil {
		return err
	}
	defer publicFd.Close()
	err = pem.Encode(publicFd, block)
	if err != nil {
		return err
	}
	return nil
}
