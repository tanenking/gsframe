package gsframe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash/fnv"
	"io"

	"github.com/tanenking/gsframe/internal/logx"
)

func GetHash32s(s string) uint32 {
	return GetHash32([]byte(s))
}

func GetHash32(b []byte) uint32 {
	h32 := fnv.New32()
	h32.Write(b)
	return h32.Sum32()
}
func GetHash64s(s string) uint64 {
	return GetHash64([]byte(s))
}
func GetHash64(b []byte) uint64 {
	h64 := fnv.New64()
	h64.Write(b)
	return h64.Sum64()
}

func GetHmacSha1(data, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func GetHmacSha256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))

	return hex.EncodeToString(mac.Sum(nil))
}

func GetMD5(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}
func GCMEncrypt(text, secretKey string) (string, error) {
	key, _ := hex.DecodeString(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.ErrorF("GCMEncrypt NewCipher err = %v", err)
		return "", err
	}
	aeaGcm, err := cipher.NewGCM(block)
	if err != nil {
		logx.ErrorF("GCMEncrypt NewGCM err = %v", err)
		return "", err
	}
	nonce := make([]byte, aeaGcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logx.ErrorF("GCMEncrypt ReadFull err = %v", err)
		return "", err
	}
	cipherText := aeaGcm.Seal(nonce, nonce, []byte(text), nil)
	encoded := base64.StdEncoding.EncodeToString(cipherText)

	return encoded, nil
}

func EncryptoAES(str string, _key_aes string) string {
	data := []byte(str)
	if len(data) <= 0 {
		return ""
	}
	block, err := aes.NewCipher([]byte(_key_aes))
	if err != nil {
		return ""
	}
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func DecryptoAES(data string, _key_aes string) string {
	if len(data) <= 0 {
		return ""
	}
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher([]byte(_key_aes))
	if err != nil {
		return ""
	}
	if len(ciphertext) < aes.BlockSize {
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}
