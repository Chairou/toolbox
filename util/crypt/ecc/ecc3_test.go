package ecc

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"

	//以太坊加密库，要求go版本升级到1.15
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func genPrivateKey() (*ecies.PrivateKey, error) {
	pubkeyCurve := crypto.S256() //初始化椭圆曲线
	//随机挑选基点，生成私钥
	p, err := ecdsa.GenerateKey(pubkeyCurve, rand.Reader) //用golang标准库生成公私钥
	if err != nil {
		return nil, err
	} else {
		return ecies.ImportECDSA(p), nil //转换成以太坊的公私钥对
	}
}

// ECCEncrypt2 椭圆曲线加密
func ECCEncrypt(plain string, pubKey *ecies.PublicKey) ([]byte, error) {
	src := []byte(plain)
	return ecies.Encrypt(rand.Reader, pubKey, src, nil, nil)
}

// ECCDecrypt2 椭圆曲线解密
func ECCDecrypt(cipher []byte, prvKey *ecies.PrivateKey) (string, error) {
	if src, err := prvKey.Decrypt(cipher, nil, nil); err != nil {
		return "", err
	} else {
		return string(src), nil
	}
}

func TestEcc3(t *testing.T) {
	prvKey, err := genPrivateKey()
	if err != nil {
		fmt.Println(err)
	}
	pubKey := prvKey.PublicKey
	plain := "我们没什么不同"
	cipher, err := ECCEncrypt(plain, &pubKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("密文：%v\n", cipher)
	plain, err = ECCDecrypt(cipher, prvKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("明文：%s\n", plain)
}
