package crypt

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const DefaultKey string = "!7@8#0$9%5^1&3*2(4)t6yKfqngDfg,."

func AesEncrypt2(origData []byte, key string) string {
	tmpKey := key + DefaultKey
	realKey := tmpKey[:32]

	// 转成字节数组
	var k []byte
	if key != "" {
		k = []byte(realKey)
	} else {
		k = []byte(DefaultKey)
	}

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		fmt.Println(err)
		return ""
	}
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

	return base64.StdEncoding.EncodeToString(cryted)

}

func AesDecrypt2(cryted string, key string) []byte {
	tmpKey := key + "12345678901234567890123456789012"
	realKey := tmpKey[:32]
	// 转成字节数组
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var k []byte
	if key != "" {
		k = []byte(realKey)
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
	return orig
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

// EncryptLargeFile 加密大文件
func EncryptLargeFile(key []byte, inputPath, outputPath string) error {
	tmpKey := append(key, []byte(DefaultKey)...)
	realKey := tmpKey[:32]
	input, _ := os.Open(inputPath)
	defer func(input *os.File) {
		err := input.Close()
		if err != nil {
			fmt.Println("EncryptLargeFile|os.Open error", err)
		}
	}(input)
	output, _ := os.Create(outputPath)
	defer func(output *os.File) {
		err := output.Close()
		if err != nil {
			fmt.Println("EncryptLargeFile|os.Create error", err)
		}
	}(output)

	// 生成随机 IV 并写入文件头部
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		return err
	}
	_, err = output.Write(iv)
	if err != nil {
		fmt.Println("EncryptLargeFile |Write iv error", err)
		return err
	}

	block, _ := aes.NewCipher(realKey)
	encrypter := cipher.NewCBCEncrypter(block, iv)
	reader := bufio.NewReaderSize(input, 64*1024*1024) // 64MB 缓冲

	for {
		buf := make([]byte, reader.Size())
		n, err := reader.Read(buf)
		if n > 0 {
			// 仅对最后一块进行填充
			if n%aes.BlockSize != 0 {
				padded := PKCS7Padding(buf[:n], aes.BlockSize)
				encrypter.CryptBlocks(buf, padded)
				_, err = output.Write(buf[:len(padded)])
			} else {
				encrypter.CryptBlocks(buf, buf)
				_, err = output.Write(buf)
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("EncryptLargeFile|Read error", err)
			return err
		}
	}
	return nil
}

// DecryptLargeFile 解密大文件
func DecryptLargeFile(key []byte, inputPath, outputPath string) error {
	tmpKey := append(key, []byte(DefaultKey)...)
	realKey := tmpKey[:32]
	input, _ := os.Open(inputPath)
	defer func(input *os.File) {
		err := input.Close()
		if err != nil {
			fmt.Println("DecryptLargeFile|os.Open error", err)
		}
	}(input)
	output, _ := os.Create(outputPath)
	defer func(output *os.File) {
		err := output.Close()
		if err != nil {
			fmt.Println("DecryptLargeFile|os.Create error", err)
		}
	}(output)

	// 从文件头部读取 IV
	iv := make([]byte, aes.BlockSize)
	_, err := input.Read(iv)
	if err != nil {
		fmt.Println("DecryptLargeFile|Read iv error", err)
		return err
	}

	block, _ := aes.NewCipher(realKey)
	decrypter := cipher.NewCBCDecrypter(block, iv)
	reader := bufio.NewReaderSize(input, 64*1024*1024)

	for {
		buf := make([]byte, reader.Size())
		n, err := reader.Read(buf)
		if n <= reader.Size() {
			decrypter.CryptBlocks(buf, buf[:n])
			// 仅在文件末尾移除填充
			if n < reader.Size() {
				buf = PKCS7UnPadding(buf[:n])
				output.Write(buf)
				return nil
			}
			_, err2 := output.Write(buf[:n])
			if err2 != nil {
				fmt.Println("DecryptLargeFile|Write buf error", err2)
				return err2
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("DecryptLargeFile|Read error", err)
			return err
		}
	}
	return nil
}
