package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

var DefaultKey string = "0987654321!@#$%^&*()tyKfqngDfg,."

func AesEncrypt2(orig string, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes")
	}
	// 转成字节数组
	origData := []byte(orig)
	var k []byte
	if key != "" {
		k = []byte(key)
	} else {
		k = []byte(DefaultKey)
	}

	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return base64.StdEncoding.EncodeToString(cryted), nil

}

func AesDecrypt2(cryted string, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes")
	}
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	var k []byte
	if key != "" {
		k = []byte(key)
	} else {
		k = []byte(DefaultKey)
	}
	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		fmt.Println(err)
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig), nil
}

// PKCS7Padding 补码
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
