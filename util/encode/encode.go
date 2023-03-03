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
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func FlatCompress(origData []byte) (result []byte, err error) {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, -1)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(origData)
	err = w.Close()
	result = buf.Bytes()
	return
}

func FlatUnCompress(compressData []byte) (result []byte, err error) {
	result, err = io.ReadAll(flate.NewReader(bytes.NewReader(compressData)))
	return
}

func Base64Encode(origData []byte) (result string) {
	result = base64.StdEncoding.EncodeToString(origData)
	return result
}

func Base64Decode(encodedData string) (result []byte, err error) {
	result, err = base64.StdEncoding.DecodeString(encodedData)
	return result, err
}

func MD5(origData []byte) (result string) {
	has := md5.Sum(origData)
	result = fmt.Sprintf("%x", has) //将[]byte转成16进制
	return result
}

func MD5File(fileName string) (result string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("MD5File|Close err:", err)
		}
	}(f)

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	result = hex.EncodeToString(h.Sum(nil))
	return result, nil
}

func Sha1(origData []byte) (result string) {
	h := sha1.New()
	h.Write(origData)
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result
}

func Sha1File(fileName string) (result string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("MD5File|Close err:", err)
		}
	}(f)

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result, nil
}

func Sha256(origData []byte) (result string) {
	h := sha256.New()
	h.Write(origData)
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result
}

func Sha256File(fileName string) (result string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("MD5File|Close err:", err)
		}
	}(f)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result, nil
}

func Sha512(origData []byte) (result string) {
	h := sha512.New()
	h.Write(origData)
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result
}

func Sha512File(fileName string) (result string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("MD5File|Close err:", err)
		}
	}(f)
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	result = fmt.Sprintf("%x", h.Sum(nil))
	return result, nil
}

// PKCS5Padding @brief:填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, paddingText...)
}

// PKCS5UnPadding @brief:去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AesEncrypt @brief:AES加密
// AES秘钥的长度只能是16、24或32字节，分别对应三种AES，即AES-128, AES-192和AES-256，三者的区别是加密的轮数不同；
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// AesDecrypt @brief:AES解密
// AES秘钥的长度只能是16、24或32字节，分别对应三种AES，即AES-128, AES-192和AES-256，三者的区别是加密的轮数不同；
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}
