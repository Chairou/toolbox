package ecc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"math/big"
	"os"
	"runtime"
)

var DesKeyError = "key size should be 8"
var DesIvError = "IV size should be 8"
var TripleDesKeyError = "key size should be 24"
var AesKeyError = "key size should be 16, 24 or 32"
var AesIvError = "IV size should be 16, 24 or 32"
var RsatransError = "error occur when trans to *rsa.Publickey"

// var RsaNilError = "error occur when decrypt"
var EcckeyError = "key size should be 224 256 384 521"

// 生成ECC私钥对
// keySize 密钥大小, 224 256 384 521
// dirPath 密钥文件生成后保存的目录
// 返回 错误
func GenerateECCKey(keySize int, dirPath string) error {
	// generate private key
	var priKey *ecdsa.PrivateKey
	var err error
	switch keySize {
	case 224:
		priKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case 256:
		priKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case 384:
		priKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case 521:
		priKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		priKey, err = nil, nil
	}
	if priKey == nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, EcckeyError)
	}
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	// x509
	derText, err := x509.MarshalECPrivateKey(priKey)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	// pem block
	block := &pem.Block{
		Type:  "ecdsa private key",
		Bytes: derText,
	}
	file, err := os.Create(dirPath + "eccPrivate.pem")
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	err = pem.Encode(file, block)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	file.Close()
	// public key
	pubKey := priKey.PublicKey
	derText, err = x509.MarshalPKIXPublicKey(&pubKey)
	block = &pem.Block{
		Type:  "ecdsa public key",
		Bytes: derText,
	}
	file, err = os.Create(dirPath + "eccPublic.pem")
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	err = pem.Encode(file, block)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return Error(file, line+1, err.Error())
	}
	file.Close()
	return nil
}

// Ecc 加密
// plainText 明文
// filePath 公钥文件路径
// 返回 密文 错误
func EccEncrypt(plainText []byte, filePath string) ([]byte, error) {
	// get pem.Block
	block, err := GetKey(filePath)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	// X509
	publicInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	publicKey, flag := publicInterface.(*ecdsa.PublicKey)
	if flag == false {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, RsatransError)
	}
	cipherText, err := ecies.Encrypt(rand.Reader, PubEcdsaToEcies(publicKey), plainText, nil, nil)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	return cipherText, err
}

// ECC 解密
// cipherText 密文
// filePath 私钥文件路径
// 返回 明文 错误
func EccDecrypt(cipherText []byte, filePath string) (plainText []byte, err error) {
	// get pem.Block
	block, err := GetKey(filePath)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	// get privateKey
	privateKey, _ := x509.ParseECPrivateKey(block.Bytes)
	priKey := PriEcdsaToEcies(privateKey)
	plainText, err = priKey.Decrypt(cipherText, nil, nil)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	return plainText, nil
}

// Ecc 签名
// plainText 明文
// priPath 私钥路径
// 返回 签名结果
func ECCSign(plainText []byte, priPath string) ([]byte, []byte, error) {
	// get pem.Block
	block, err := GetKey(priPath)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, nil, Error(file, line+1, err.Error())
	}
	// x509
	priKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, nil, Error(file, line+1, err.Error())
	}
	hashText := sha256.Sum256(plainText)
	// sign
	r, s, err := ecdsa.Sign(rand.Reader, priKey, hashText[:])
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, nil, Error(file, line+1, err.Error())
	}
	// marshal
	rText, err := r.MarshalText()
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, nil, Error(file, line+1, err.Error())
	}
	sText, err := s.MarshalText()
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, nil, Error(file, line+1, err.Error())
	}
	return rText, sText, nil
}

// ECC 签名验证
// plainText 明文
// rText,sText 签名
// pubPath公钥文件路径
// 返回 验签结果 错误
func ECCVerify(plainText, rText, sText []byte, pubPath string) (bool, error) {
	// get pem.Block
	block, err := GetKey(pubPath)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return false, Error(file, line+1, err.Error())
	}
	// x509
	pubInter, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return false, Error(file, line+1, err.Error())
	}
	// assert
	pubKey := pubInter.(*ecdsa.PublicKey)
	hashText := sha256.Sum256(plainText)
	var r, s big.Int
	// unmarshal
	err = r.UnmarshalText(rText)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return false, Error(file, line+1, err.Error())
	}
	err = s.UnmarshalText(sText)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return false, Error(file, line+1, err.Error())
	}
	// verify
	ok := ecdsa.Verify(pubKey, hashText[:], &r, &s)
	return ok, nil
}

// 填充最后一组
// plainText：明文
// blockSize：块大小
// 返回：填充后的明文
func PaddingLastGroup(plainText []byte, blockSize int) []byte {
	// get size num of last group, eg 28%8 = 4, 32%8=0
	padNum := blockSize - len(plainText)%blockSize
	// create a new slice, length equals padNum
	char := []byte{byte(padNum)}
	newPlain := bytes.Repeat(char, padNum)
	// append newPlain to plainText
	plainText = append(plainText, newPlain...)
	return plainText
}

// 去掉填充
// plainText：填充后的明文
// 返回：填充前的明文
func UnpaddingLastGroup(plainText []byte) []byte {
	length := len(plainText)
	// get length we need to cut
	number := int(plainText[length-1])
	return plainText[:length-number]
}

// 错误格式化
func Error(file string, line int, err string) error {
	return fmt.Errorf("file:%s line:%d error:%s", file, line, err)
}

// 读取公钥/私钥文件，获取解码的pem块
// filePath文件路径
// 返回pem块和错误
func GetKey(filePath string) (*pem.Block, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	fileInfo, err := file.Stat()
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	buf := make([]byte, fileInfo.Size())
	_, err = file.Read(buf)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, Error(file, line+1, err.Error())
	}
	block, _ := pem.Decode(buf)
	return block, err
}

// ecdsa public key to ecies public key
func PubEcdsaToEcies(pub *ecdsa.PublicKey) *ecies.PublicKey {
	return &ecies.PublicKey{
		X:      pub.X,
		Y:      pub.Y,
		Curve:  pub.Curve,
		Params: ecies.ParamsFromCurve(pub.Curve),
	}
}

// ecdsa private key to ecies private key
func PriEcdsaToEcies(prv *ecdsa.PrivateKey) *ecies.PrivateKey {
	pub := PubEcdsaToEcies(&prv.PublicKey)
	return &ecies.PrivateKey{*pub, prv.D}
}
