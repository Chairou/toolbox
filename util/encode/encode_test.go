package encode

import (
	"fmt"
	"testing"
)

func TestAesEnDeCrypt(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	crypted, _ := AesEncrypt([]byte(str), []byte("12345678901234567890123456789012"))
	decryptBytes, _ := AesDecrypt(crypted, []byte("12345678901234567890123456789012"))
	if string(decryptBytes) != str {
		t.Error("must be string 我是努力的八王子ABCabc123!@#, not string", string(decryptBytes))
	}
}

func TestBase64EnDecode(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	b64Str := Base64Encode([]byte(str))
	retStr, _ := Base64Decode(b64Str)
	if string(retStr) != str {
		t.Error("must be string 我是努力的八王子ABCabc123!@#, not string", string(retStr))
	}
}

func TestFlateCompress(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	compressBytes, _ := FlateCompress([]byte(str))
	deCompressBytes, _ := FlateUnCompress(compressBytes)
	if string(deCompressBytes) != str {
		t.Error("must be string 我是努力的八王子ABCabc123!@#, not string", string(deCompressBytes))
	}
}

func TestMD5(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	md5Bytes := MD5([]byte(str))
	if md5Bytes != "98ee2def518eb939ac2c3c81716018d5" {
		t.Error("must be string 41fd09f81e06d2fad2f5e7f3403574e2, not string", md5Bytes)

	}
}

func TestSha1(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha1Str := Sha1([]byte(str))
	if sha1Str != "d7378d69ac75bf9d04c5d9742d068fedd01b6e72" {
		t.Error("must be string d7378d69ac75bf9d04c5d9742d068fedd01b6e72, not string", (sha1Str))

	}
}

func TestSha256(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha256Str := Sha256([]byte(str))
	if sha256Str != "09f5e6dbe4bb0ca21f62bb26c5916b043aa2481547c14c7df7d79383421c6a36" {
		t.Error("must be string 09f5e6dbe4bb0ca21f62bb26c5916b043aa2481547c14c7df7d79383421c6a36, "+
			"not string", sha256Str)
	}
}

func TestSha512(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha512Str := Sha512([]byte(str))
	if sha512Str != "51bf8273e87226958f726da816df56fa28d0336acd1b5fd46c3570040c9c24ef2dfdb25d9"+
		"b1fed35b22e0a3a8669c8c43b92706cf8dbd870adee649144937f25" {
		t.Error("must be string 51bf8273e87226958f726da816df56fa28d0336acd1b5fd46c3570040c9c24ef2dfdb25d9"+
			"b1fed35b22e0a3a8669c8c43b92706cf8dbd870adee649144937f25, "+
			"not string", sha512Str)
	}
}

// go test -bench=.
func BenchmarkFlateCompress(b *testing.B) {
	str := "我是努力的八王子ABCabc123!@#"
	var compressBytes []byte
	fmt.Println(string(compressBytes))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compressBytes, _ = FlateCompress([]byte(str))
	}
	b.ReportAllocs()
}

func BenchmarkAesEncrypt(b *testing.B) {
	str := "我是努力的八王子ABCabc123!@#"
	var crypted []byte
	fmt.Println(string(crypted))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypted, _ = AesEncrypt([]byte(str), []byte("12345678901234567890123456789012"))
	}
	b.ReportAllocs()
}

func BenchmarkAesDecrypt(b *testing.B) {
	b.ReportAllocs()
	str := "我是努力的八王子ABCabc123!@#"
	crypted, _ := AesEncrypt([]byte(str), []byte("12345678901234567890123456789012"))
	tmp := []byte{0}
	fmt.Println(string(tmp))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp, _ = AesDecrypt(crypted, []byte("12345678901234567890123456789012"))
	}
}
