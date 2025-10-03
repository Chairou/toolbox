package crypt

import (
	"testing"
)

func TestCrypt(t *testing.T) {
	// key必须是32字节或者空字符
	key := "12345678901234567890123456789012"
	//key := "asdasd"
	orig := []byte("hand in hand we stand.")
	cryptContent := AesEncrypt2(orig, key)

	output := AesDecrypt2(cryptContent, key)
	t.Logf(string(output))
}

func TestEncryptLargeFile(t *testing.T) {
	err := EncryptLargeFile([]byte("12345678901234567890123456789012"), "sample.txt", "sample.enc")
	if err != nil {
		t.Error(err)
	}
	err = DecryptLargeFile([]byte("12345678901234567890123456789012"), "sample.enc", "sample.dec")
	if err != nil {
		t.Error(err)
	}
}
