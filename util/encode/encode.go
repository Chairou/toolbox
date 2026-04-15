// Package encode 提供常用的编码、哈希、压缩和加解密工具函数
package encode

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"os"
)

// FlatCompress 使用flate算法压缩数据
func FlatCompress(origData []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, -1)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(origData)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FlatUnCompress 使用flate算法解压数据
func FlatUnCompress(compressData []byte) ([]byte, error) {
	return io.ReadAll(flate.NewReader(bytes.NewReader(compressData)))
}

// Base64Encode 对数据进行Base64编码
func Base64Encode(origData []byte) string {
	return base64.StdEncoding.EncodeToString(origData)
}

// Base64Decode 对Base64编码的字符串进行解码
func Base64Decode(encodedData string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedData)
}

// MD5 计算数据的MD5哈希值
func MD5(origData []byte) string {
	has := md5.Sum(origData)
	return fmt.Sprintf("%x", has)
}

// hashFile 通用的文件哈希计算辅助函数
func hashFile(fileName string, h hash.Hash) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// MD5File 计算文件的MD5哈希值
func MD5File(fileName string) (string, error) {
	return hashFile(fileName, md5.New())
}

// Sha1 计算数据的SHA1哈希值
func Sha1(origData []byte) string {
	h := sha1.New()
	h.Write(origData)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Sha1File 计算文件的SHA1哈希值
func Sha1File(fileName string) (string, error) {
	return hashFile(fileName, sha1.New())
}

// Sha256 计算数据的SHA256哈希值
func Sha256(origData []byte) string {
	h := sha256.New()
	h.Write(origData)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Sha256File 计算文件的SHA256哈希值
func Sha256File(fileName string) (string, error) {
	return hashFile(fileName, sha256.New())
}

// Sha512 计算数据的SHA512哈希值
func Sha512(origData []byte) string {
	h := sha512.New()
	h.Write(origData)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Sha512File 计算文件的SHA512哈希值
func Sha512File(fileName string) (string, error) {
	return hashFile(fileName, sha512.New())
}

// PKCS5Padding 对明文进行PKCS5填充
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, paddingText...)
}

// PKCS5UnPadding 去除PKCS5填充数据
func PKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, fmt.Errorf("PKCS5UnPadding: input data is empty")
	}
	unpadding := int(origData[length-1])
	if unpadding > length || unpadding == 0 {
		return nil, fmt.Errorf("PKCS5UnPadding: invalid padding size %d", unpadding)
	}
	// 校验所有填充字节是否一致
	for i := length - unpadding; i < length; i++ {
		if origData[i] != byte(unpadding) {
			return nil, fmt.Errorf("PKCS5UnPadding: invalid padding byte at position %d", i)
		}
	}
	return origData[:(length - unpadding)], nil
}

// AesEncrypt AES-CBC加密
// AES秘钥的长度只能是16、24或32字节，分别对应AES-128, AES-192和AES-256
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// AesDecrypt AES-CBC解密
// AES秘钥的长度只能是16、24或32字节，分别对应AES-128, AES-192和AES-256
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	if len(crypted) == 0 {
		return nil, fmt.Errorf("AesDecrypt: encrypted data is empty")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	if len(crypted)%blockSize != 0 {
		return nil, fmt.Errorf("AesDecrypt: encrypted data length %d is not a multiple of block size %d",
			len(crypted), blockSize)
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	return PKCS5UnPadding(origData)
}