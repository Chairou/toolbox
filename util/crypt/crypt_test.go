package crypt

import (
	"testing"
)

func TestCrypt(t *testing.T) {
	// key必须是32字节或者空字符
	key := "12345678901234567890123456789012"
	//key := "asdasd"
	orig := "hand in hand we stand."
	cryptContent, err := AesEncrypt2(orig, key)
	if err != nil {
		t.Error(err)
	}
	output, err := AesDecrypt2(cryptContent, key)
	if err != nil {
		t.Error(err)
	}
	t.Logf(output)
}
