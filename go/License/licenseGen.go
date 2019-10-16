package main

import (
	"crypto"
	"crypto/rand"
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

type LicenseGenConfig struct {
	User    string `json:"User,omitempty"`
	Version string `json:"Version,omitempty"`
	UUID    string `json:"UUID,omitempty"` // 用户设备标识
	Expire  string `json:"Expire,omitempty"`
	Message string `json:"Message,omitempty"`

	Base64Signature string `json:"Signature,omitempty"` // 用以上字段生成
}

type LicenseGen struct {
	Config LicenseGenConfig

	PrivateKeyFile string
	PrivateKey     []byte
}

func NewLicenseSign(config LicenseGenConfig, privateKeyFile string) *LicenseSign {

	// 读取私钥
	privateKey, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil
	}

	// LicenseSign构造
	return &LicenseSign{Config: config, PrivateKeyFile: privateKeyFile, PrivateKey: privateKey}
}

/* 私钥签名 */
func (l *LicenseSign) Signature(sourceMsg string) error {
	data := []byte(sourceMsg)
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)

	// 获取私钥
	block, _ := pem.Decode(l.PrivateKey)
	if block == nil {
		return errors.New("private key error")
	}

	// 解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	sign, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
	if err != nil {
		return err
	}

	// signature编码
	encodeString := base64.StdEncoding.EncodeToString(sign)
	l.Config.Base64Signature = encodeString
	return nil
}

func main() {

	var privateKeyFile string
	privateKeyFile = "private.pem"

	// License 解析
	data, err := ioutil.ReadFile("./License")
	if err != nil {
		fmt.Println("read 'License' error:", err.Error())
		return
	}
	var config LicenseGenConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Unmarshal error:", err.Error())
		return
	}
	license := NewLicenseSign(config, privateKeyFile)
	if license == nil {
		fmt.Println("NewLicenseSign error")
		return
	}

	// 私钥签名
	sourceMsg := strings.Join([]string{config.User, config.Version, config.UUID, config.Expire, config.Message}, ",")
	err = license.Signature(sourceMsg)
	if err != nil {
		fmt.Println("Signature:", err.Error())
		return
	}
	config.Base64Signature = license.Config.Base64Signature
	fmt.Println("Signature:", config.Base64Signature)

	// Licence 生成
	licenseJson, err := json.Marshal(config)
	if err != nil {
		fmt.Println("LicenseGen error:", err.Error())
		return
	}
	err = ioutil.WriteFile("License", licenseJson, 0644)
	if err != nil {
		fmt.Println("LicenseGen error:", err.Error())
		return
	}
	fmt.Println("LicenseGen 'License' OK!")
}
