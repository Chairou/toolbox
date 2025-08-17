package ecc

import (
	"fmt"
	"github.com/Chairou/toolbox/util/encode"
	"io"
	"net/url"
	"os"
)

func EncryptByteFromEccFile(filepath string, plaint []byte) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("EncryptByteFromEccFile|failed to close file: %v\n", err)
		}
	}(file)

	buf := make([]byte, 65535) // 更合理的缓冲区大小
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	pub, err := DecodePEMPublicKey(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}
	// use public key to encrypt
	return pub.Encrypt(plaint)
}

func EncryptStringFromEccFile(filepath string, plaintext string) (string, error) {
	enBytes, err := EncryptByteFromEccFile(filepath, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}
	return url.QueryEscape(encode.Base64Encode(enBytes)), nil
}

func DecryptByteFromEccFile(filepath string, passwd string, ciphertext []byte) ([]byte, error) {
	privateFile, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("DecryptByteFromEccFile|failed to close file: %v\n", err)
		}
	}(privateFile)

	buf := make([]byte, 65535) // 更合理的缓冲区大小
	n, err := privateFile.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	priv, err := DecodePEMPrivateKey(buf[:n], passwd)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %v", err)
	}

	return priv.Decrypt(ciphertext)
}

func DecryptStringFromEccFile(filepath string, passwd string, ciphertext string) (string, error) {
	queryUnescapeStr, err := url.QueryUnescape(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to unescape ciphertext: %v", err)
	}

	unb64Encrypt, err := encode.Base64Decode(queryUnescapeStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %v", err)
	}

	decrypt, err := DecryptByteFromEccFile(filepath, passwd, unb64Encrypt)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(decrypt), nil
}
