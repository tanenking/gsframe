package gsframe

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/tanenking/gsframe/internal/logx"
)

const (
	// 私钥 PEMBEGIN 开头
	PEMBEGIN = "-----BEGIN RSA PRIVATE KEY-----\n"
	// 私钥 PEMEND 结尾
	PEMEND = "\n-----END RSA PRIVATE KEY-----"
	// 公钥 PEMBEGIN 开头
	PUBPEMBEGIN = "-----BEGIN PUBLIC KEY-----\n"
	// 公钥 PEMEND 结尾
	PUBPEMEND = "\n-----END PUBLIC KEY-----"
)

// Rsa2Sign RSA2私钥签名
func Rsa2SignFromKey(signContent string, privateKey *rsa.PrivateKey, hash crypto.Hash) string {
	if privateKey == nil {
		return ""
	}
	shaNew := hash.New()
	shaNew.Write([]byte(signContent))
	hashed := shaNew.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(signature)
}

// Rsa2Sign RSA2私钥签名
func Rsa2Sign(signContent string, privateKey string, hash crypto.Hash) string {
	shaNew := hash.New()
	shaNew.Write([]byte(signContent))
	hashed := shaNew.Sum(nil)
	priKey, err := ParsePrivateKey(privateKey)
	if err != nil {
		return ""
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, priKey, hash, hashed)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(signature)
}

// ParsePrivateKey 私钥验证
func ParsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = FormatPrivateKey(privateKey)
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("私钥信息错误！")
	}
	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priKey, nil
}

// FormatPrivateKey 组装私钥
func FormatPrivateKey(privateKey string) string {
	if !strings.HasPrefix(privateKey, PEMBEGIN) {
		privateKey = PEMBEGIN + privateKey
	}
	if !strings.HasSuffix(privateKey, PEMEND) {
		privateKey = privateKey + PEMEND
	}
	return privateKey
}

// Rsa2PubSign RSA2公钥验证签名
func Rsa2PubSign(signContent, sign, publicKey string, hash crypto.Hash) bool {
	hashed := sha256.Sum256([]byte(signContent))
	pubKey, err := ParsePublicKey(publicKey)
	if err != nil {
		logx.ErrorF("rsa2 public check sign failed. err = %v", err)
		return false
	}
	sig, _ := base64.StdEncoding.DecodeString(sign)
	err = rsa.VerifyPKCS1v15(pubKey, hash, hashed[:], sig)
	if err != nil {
		logx.ErrorF("rsa2 public check sign failed. err = %v", err)
		return false
	}
	return true
}

// ParsePublicKey 公钥验证
func ParsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	publicKey = FormatPublicKey(publicKey)
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("公钥信息错误！")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubKey.(*rsa.PublicKey), nil
}

// FormatPublicKey 组装公钥
func FormatPublicKey(publicKey string) string {
	if !strings.HasPrefix(publicKey, PUBPEMBEGIN) {
		publicKey = PUBPEMBEGIN + publicKey
	}
	if !strings.HasSuffix(publicKey, PUBPEMEND) {
		publicKey = publicKey + PUBPEMEND
	}
	return publicKey
}
