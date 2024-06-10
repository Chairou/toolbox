package ecc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"strconv"
	"testing"
)

// 获取私钥
func getKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return privateKey, err
	}
	return privateKey, nil
}

// 加密
func ECCEncrypt2(publicKey ecies.PublicKey, data []byte) ([]byte, error) {
	ct, err := ecies.Encrypt(rand.Reader, &publicKey, data, nil, nil)
	return ct, err
}

// 解密
func ECCDecrypt2(privateKey ecies.PrivateKey, ct []byte) ([]byte, error) {
	m, err := privateKey.Decrypt(ct, nil, nil)
	return m, err
}

// 获取哈希
func getHash2(data []byte, nonce int) string {
	hashBytes := sha256.Sum256([]byte(string(data) + strconv.Itoa(nonce)))
	return hex.EncodeToString(hashBytes[:])
}

// 获取挖矿等级需求来进行判断hash是否满足
func getMineDiff(diff int) (str string) {
	for i := 0; i < diff; i++ {
		str = str + "0"
	}
	return
}

// 开始挖矿[链的挖矿难度]
func calculationHash(diff int, data []byte) string {
	strDiff := getMineDiff(diff)
	var nonce int
	for {
		if getHash2(data, nonce)[:diff] == strDiff {
			return getHash2(data, nonce)
		}
		nonce++
	}
}

// 签名
func signature(hash []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return signature, err
	}
	return signature, nil
}

// 验证
func validate(recoveredPub, recoveredPubBytes []byte) bool {
	if !bytes.Equal(recoveredPubBytes, recoveredPub) {
		return false
	}
	return true
}

func TestEcc4(t *testing.T) {
	privateKeyECDSA, _ := getKey()
	// 将ecdsa的私钥转换成以太坊的私钥
	privateKey := ecies.ImportECDSA(privateKeyECDSA)
	publicKey := privateKey.PublicKey
	// 公钥加密
	data := []byte("我向某用户转账10元")
	hash := calculationHash(4, data)

	fmt.Println("哈希散列为", hash)
	encryptData, err := ECCEncrypt2(publicKey, []byte(data))
	if err != nil {
		panic(err)
	}
	fmt.Println("公钥加密后为", hex.EncodeToString(encryptData))

	// 私钥解密
	decryptData, err := ECCDecrypt2(*privateKey, encryptData)
	if err != nil {
		panic(err)
	}
	fmt.Println("私钥解密后为", string(decryptData))

	// 进行签名
	signData, _ := signature(crypto.Keccak256([]byte(hash)), privateKey.ExportECDSA())
	fmt.Println("签名为", hex.EncodeToString(signData))

	// 验证
	recoveredPub, _ := crypto.Ecrecover(crypto.Keccak256([]byte(hash)), signData)
	recoveredPubBytes := crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)

	fmt.Println(validate(recoveredPub, recoveredPubBytes))
}
