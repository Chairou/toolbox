package ecc

import (
	"github.com/Chairou/toolbox/util/encode"
	"net/url"
	"os"
	"testing"
)

//func TestCreatePem(t *testing.T) {
//	pub, priv, err := GenerateKeys(elliptic.P521())
//	if err != nil {
//		panic(err)
//	}
//	publicKeyFile, err := os.Create("public.pem")
//	if err != nil {
//		t.Error(err)
//	}
//	defer publicKeyFile.Close()
//	bytes, err := pub.PEM()
//	if err != nil {
//		panic(err)
//	}
//	_, err = publicKeyFile.Write(bytes)
//	if err != nil {
//		panic(err)
//	}
//	privateKeyFile, err := os.Create("private.pem")
//	if err != nil {
//		panic(err)
//	}
//	bytes, err = priv.PEM("123")
//	if err != nil {
//		panic(err)
//	}
//	_, err = privateKeyFile.Write(bytes)
//	if err != nil {
//		panic(err)
//	}
//}

//func TestEcc5(t *testing.T) {
//	// create keys
//	pub, priv, _ := GenerateKeys(elliptic.P521())
//
//	plaintext := "secret secrets are no fun, secret secrets hurt someone"
//
//	// use public key to encrypt
//	encrypted, _ := pub.Encrypt([]byte(plaintext))
//	fmt.Println(len(encrypted))
//
//	// use private key to decrypt
//	decrypted, _ := priv.Decrypt(encrypted)
//	fmt.Println(string(decrypted))
//
//}

func TestEccFromFile(t *testing.T) {
	file, _ := os.Open("public.pem")
	//file, _ := os.Open("2aes.pem")
	buf := make([]byte, 65536)
	_, _ = file.Read(buf)
	pub, err := DecodePEMPublicKey(buf)
	if err != nil {
		t.Error(err)
	}
	plaintext := "secret secrets are no fun, secret secrets hurt someone!"

	// use public key to encrypt
	encrypted, _ := pub.Encrypt([]byte(plaintext))
	b64ecrypt := encode.Base64Encode(encrypted)
	queryEscapeStr := url.QueryEscape(b64ecrypt)
	t.Log("enStr:", queryEscapeStr)

	prifile, _ := os.Open("private.pem")
	//prifile, _ := os.Open("1.pem")
	_, _ = prifile.Read(buf)
	queryUnescapeStr, _ := url.QueryUnescape(queryEscapeStr)
	unb64Encrypt, _ := encode.Base64Decode(queryUnescapeStr)
	priv, err := DecodePEMPrivateKey(buf, "123")
	decrypt, err := priv.Decrypt(unb64Encrypt)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(decrypt))
}

func TestEccFromFunction(t *testing.T) {
	enStr, err := EncryptStringFromEccFile("public.pem", "hand in 2hand we stand")
	if err != nil {
		t.Error(err)
	}
	plaint, err := DecryptStringFromEccFile("1.pem", "123", enStr)
	if err != nil {
		t.Error(err)
	}
	t.Log(plaint)
}
