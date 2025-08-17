// Package symcrypt contains common symmetric encryption functions.
package symcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"math/big"
)

// DecryptCBC decrypt bytes using a key and IV with AES in CBC mode.
func DecryptCBC(data, iv, key []byte) (decryptedData []byte, err error) {
	aesCrypt, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ivBytes := append([]byte{}, iv...)

	decryptedData = make([]byte, len(data))
	aesCBC := cipher.NewCBCDecrypter(aesCrypt, ivBytes)
	aesCBC.CryptBlocks(decryptedData, data)

	return
}

// EncryptCBC encrypt data using a key and IV with AES in CBC mode.
func EncryptCBC(data, iv, key []byte) (encryptedData []byte, err error) {
	aesCrypt, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ivBytes := append([]byte{}, iv...)

	encryptedData = make([]byte, len(data))
	aesCBC := cipher.NewCBCEncrypter(aesCrypt, ivBytes)
	aesCBC.CryptBlocks(encryptedData, data)

	return
}

// MakeRandom is a helper that makes a new buffer full of random data.
func MakeRandom(length int) ([]byte, error) {
	type randomData struct {
		bytes []byte
	}
	const MaxLen = 4096
	//randList := make([]randomData, length)
	randList := make([]randomData, MaxLen)
	Bytes := make([]byte, length)
	for k, _ := range randList {
		randList[k].bytes = make([]byte, length)
		_, err := rand.Read(randList[k].bytes)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < length; i++ {
		// s
		n, err := rand.Int(rand.Reader, big.NewInt(4294967296))
		if err != nil {
			return nil, err
		}
		Bytes[i] = randList[n.Int64()%int64(MaxLen)].bytes[i]
	}
	fmt.Println("Bytes:", Bytes)
	return Bytes, nil
}

func RandomInRange(min, max int) (int, error) {
	// 生成一个 *big.Int，它的值在 [0, max-min) 范围内
	delta := big.NewInt(int64(max - min))
	n, err := rand.Int(rand.Reader, delta)
	if err != nil {
		return 0, err
	}

	// 将生成的随机数转换为 int，并加上 min，得到 [min, max) 范围内的随机数
	return int(n.Int64()) + min, nil
}
