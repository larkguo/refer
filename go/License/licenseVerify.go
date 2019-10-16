package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type LicenseVerifyConfig struct {
	User    string `json:"User,omitempty"`
	Version string `json:"Version,omitempty"`
	UUID    string `json:"UUID,omitempty"` // 用户设备标识
	Expire  string `json:"Expire,omitempty"`
	Message string `json:"Message,omitempty"`

	Base64Signature string `json:"Signature,omitempty"` // 与用以上字段生成Base64Signature比较
}

type LicenseVerify struct {
	Config LicenseVerifyConfig

	PublicKeyFile string
	PublicKey     []byte
}

func NewLicenseVerify(config LicenseVerifyConfig, publicKeyFile string) *LicenseVerify {
	// 读取公钥
	publicKey, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		return nil
	}

	// LicenseVerify构造
	return &LicenseVerify{Config: config, PublicKeyFile: publicKeyFile, PublicKey: publicKey}
}

/* 公钥验证 */
func (l *LicenseVerify) SignatureVerify(sourceMsg string) error {
	data := []byte(sourceMsg)
	hashed := sha256.Sum256(data)
	block, _ := pem.Decode(l.PublicKey)
	if block == nil {
		return errors.New("public key error")
	}

	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)

	// base64解码
	signature, err := base64.StdEncoding.DecodeString(l.Config.Base64Signature)
	if err != nil {
		return err
	}

	// TODO: UUID 用户设备验证

	// 验证签名
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signature)
}

func main() {

	var publicKeyFile string
	publicKeyFile = "public.pem"

	// License 解析
	data, err := ioutil.ReadFile("License")
	if err != nil {
		fmt.Println("read 'License' error:", err.Error())
		return
	}
	var config LicenseVerifyConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Verify License error:", err.Error())
		return
	}
	license := NewLicenseVerify(config, publicKeyFile)
	if license == nil {
		fmt.Println("Init Verify License error!")
		return
	}

	// 公钥验证
	sourceMsg := strings.Join([]string{config.User, config.Version, config.UUID, config.Expire, config.Message}, ",")
	err = license.SignatureVerify(sourceMsg)
	if err != nil {
		fmt.Println("Verify License error:", err.Error())
	} else {
		fmt.Println("Verify License OK!")
	}
}
