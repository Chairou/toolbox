package ecc

import (
	"fmt"
	"testing"
)

func TestGenerateECCKey(t *testing.T) {
	//err := GenerateECCKey(224, "./")
	//if err != nil {
	//	fmt.Println(err)
	//}
	////os.Exit(0)
	//err = GenerateECCKey(256, "./")
	//if err != nil {
	//	fmt.Println(err)
	//}
	////os.Exit(0)
	//err = GenerateECCKey(384, "./")
	//if err != nil {
	//	fmt.Println(err)
	//}
	////os.Exit(0)
	err := GenerateECCKey(521, "./")
	if err != nil {
		fmt.Println(err)
	}
}

func TestEcc(t *testing.T) {
	plainText := []byte("hi, I'm lady_killer9")
	cipherText, err := EccEncrypt(plainText, "./eccPublic.pem")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("加密后：%s\n", cipherText)
	plainText, err = EccDecrypt(cipherText, "./eccPrivate.pem")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("解密后：%s\n", plainText)
}

func TestSignVerify(t *testing.T) {
	plainText := []byte("张华考上了北京大学；李萍进了中等技术学校；我在百货公司当售货员：我们都有美好的未来")
	rText, sText, _ := ECCSign(plainText, "./eccPrivate.pem")
	ok, err := ECCVerify(plainText, rText, sText, "./eccPublic.pem")
	fmt.Println(err)
	fmt.Printf("验证成功？ %t", ok)
}
