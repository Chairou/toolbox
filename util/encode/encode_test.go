package encode

import (
	"testing"
)

func TestAesEnDeCrypt(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	key := []byte("12345678901234567890123456789012")
	crypted, err := AesEncrypt([]byte(str), key)
	if err != nil {
		t.Fatal("AesEncrypt err:", err)
	}
	decryptBytes, err := AesDecrypt(crypted, key)
	if err != nil {
		t.Fatal("AesDecrypt err:", err)
	}
	if string(decryptBytes) != str {
		t.Errorf("AesDecrypt result = %s, want %s", string(decryptBytes), str)
	}
}

func TestAesEncrypt_InvalidKeyLength(t *testing.T) {
	str := "test"
	// 非法密钥长度（15字节）
	_, err := AesEncrypt([]byte(str), []byte("123456789012345"))
	if err == nil {
		t.Error("AesEncrypt should fail with invalid key length")
	}
}

func TestAesDecrypt_EmptyData(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	_, err := AesDecrypt([]byte{}, key)
	if err == nil {
		t.Error("AesDecrypt should fail with empty data")
	}
}

func TestAesDecrypt_InvalidBlockSize(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	// 密文长度不是blockSize的整数倍
	_, err := AesDecrypt([]byte("12345"), key)
	if err == nil {
		t.Error("AesDecrypt should fail with data not aligned to block size")
	}
}

func TestBase64EnDecode(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	b64Str := Base64Encode([]byte(str))
	retStr, err := Base64Decode(b64Str)
	if err != nil {
		t.Fatal("Base64Decode err:", err)
	}
	if string(retStr) != str {
		t.Errorf("Base64Decode result = %s, want %s", string(retStr), str)
	}
}

func TestBase64Decode_InvalidInput(t *testing.T) {
	_, err := Base64Decode("!!!invalid-base64!!!")
	if err == nil {
		t.Error("Base64Decode should fail with invalid input")
	}
}

func TestBase64_EmptyInput(t *testing.T) {
	encoded := Base64Encode([]byte{})
	decoded, err := Base64Decode(encoded)
	if err != nil {
		t.Fatal("Base64Decode err:", err)
	}
	if len(decoded) != 0 {
		t.Errorf("Base64Decode of empty input should return empty, got %v", decoded)
	}
}

func TestFlateCompress(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	compressBytes, err := FlatCompress([]byte(str))
	if err != nil {
		t.Fatal("FlatCompress err:", err)
	}
	deCompressBytes, err := FlatUnCompress(compressBytes)
	if err != nil {
		t.Fatal("FlatUnCompress err:", err)
	}
	if string(deCompressBytes) != str {
		t.Errorf("FlatUnCompress result = %s, want %s", string(deCompressBytes), str)
	}
}

func TestFlateCompress_EmptyInput(t *testing.T) {
	compressBytes, err := FlatCompress([]byte{})
	if err != nil {
		t.Fatal("FlatCompress err:", err)
	}
	deCompressBytes, err := FlatUnCompress(compressBytes)
	if err != nil {
		t.Fatal("FlatUnCompress err:", err)
	}
	if len(deCompressBytes) != 0 {
		t.Errorf("FlatUnCompress of empty input should return empty, got %v", deCompressBytes)
	}
}

func TestMD5(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	md5Str := MD5([]byte(str))
	expected := "98ee2def518eb939ac2c3c81716018d5"
	if md5Str != expected {
		t.Errorf("MD5 result = %s, want %s", md5Str, expected)
	}
}

func TestMD5_EmptyInput(t *testing.T) {
	md5Str := MD5([]byte{})
	if md5Str == "" {
		t.Error("MD5 of empty input should not return empty string")
	}
}

func TestSha1(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha1Str := Sha1([]byte(str))
	expected := "d7378d69ac75bf9d04c5d9742d068fedd01b6e72"
	if sha1Str != expected {
		t.Errorf("Sha1 result = %s, want %s", sha1Str, expected)
	}
}

func TestSha256(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha256Str := Sha256([]byte(str))
	expected := "09f5e6dbe4bb0ca21f62bb26c5916b043aa2481547c14c7df7d79383421c6a36"
	if sha256Str != expected {
		t.Errorf("Sha256 result = %s, want %s", sha256Str, expected)
	}
}

func TestSha512(t *testing.T) {
	str := "我是努力的八王子ABCabc123!@#"
	sha512Str := Sha512([]byte(str))
	expected := "51bf8273e87226958f726da816df56fa28d0336acd1b5fd46c3570040c9c24ef2dfdb25d9" +
		"b1fed35b22e0a3a8669c8c43b92706cf8dbd870adee649144937f25"
	if sha512Str != expected {
		t.Errorf("Sha512 result = %s, want %s", sha512Str, expected)
	}
}

func TestMD5File(t *testing.T) {
	md5Str, err := MD5File("testfile.txt")
	if err != nil {
		t.Fatal("MD5File err:", err)
	}
	expected := "2886c3704317e0fb012e4a002c5d2a56"
	if md5Str != expected {
		t.Errorf("MD5File result = %s, want %s", md5Str, expected)
	}
}

func TestMD5File_NotExist(t *testing.T) {
	_, err := MD5File("not_exist_file.txt")
	if err == nil {
		t.Error("MD5File should fail with non-existent file")
	}
}

func TestSha1File(t *testing.T) {
	sha1Str, err := Sha1File("testfile.txt")
	if err != nil {
		t.Fatal("Sha1File err:", err)
	}
	expected := "0eb72683588d104c2d6d62a2bb1b671e157feb0a"
	if sha1Str != expected {
		t.Errorf("Sha1File result = %s, want %s", sha1Str, expected)
	}
}

func TestSha256File(t *testing.T) {
	sha256Str, err := Sha256File("testfile.txt")
	if err != nil {
		t.Fatal("Sha256File err:", err)
	}
	expected := "3e456bcbdd4e3f32de40c7482fcca896da15161cd4cd6d8e561cafc6c3122f7f"
	if sha256Str != expected {
		t.Errorf("Sha256File result = %s, want %s", sha256Str, expected)
	}
}

func TestSha512File(t *testing.T) {
	sha512Str, err := Sha512File("testfile.txt")
	if err != nil {
		t.Fatal("Sha512File err:", err)
	}
	expected := "0611b5b4d806ec5f50c39b26e68d1458221158bc012d96561e7ab9ea09c46922b71cc09256afade818afa626328204d" +
		"a84af442762251104db49f1fcdc25994c"
	if sha512Str != expected {
		t.Errorf("Sha512File result = %s, want %s", sha512Str, expected)
	}
}

func TestPKCS5UnPadding_EmptyInput(t *testing.T) {
	_, err := PKCS5UnPadding([]byte{})
	if err == nil {
		t.Error("PKCS5UnPadding should fail with empty input")
	}
}

func TestPKCS5UnPadding_InvalidPadding(t *testing.T) {
	// padding值大于数据长度
	_, err := PKCS5UnPadding([]byte{0x10})
	if err == nil {
		t.Error("PKCS5UnPadding should fail with invalid padding size")
	}
}

func TestPKCS5Padding_Roundtrip(t *testing.T) {
	origData := []byte("hello")
	blockSize := 16
	padded := PKCS5Padding(origData, blockSize)
	if len(padded)%blockSize != 0 {
		t.Errorf("PKCS5Padding result length %d is not a multiple of block size %d", len(padded), blockSize)
	}
	unpadded, err := PKCS5UnPadding(padded)
	if err != nil {
		t.Fatal("PKCS5UnPadding err:", err)
	}
	if string(unpadded) != "hello" {
		t.Errorf("PKCS5UnPadding result = %s, want hello", string(unpadded))
	}
}

// go test -bench=.
func BenchmarkFlateCompress(b *testing.B) {
	str := "我是努力的八王子ABCabc123!@#"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FlatCompress([]byte(str))
	}
	b.ReportAllocs()
}

func BenchmarkAesEncrypt(b *testing.B) {
	str := "我是努力的八王子ABCabc123!@#"
	key := []byte("12345678901234567890123456789012")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AesEncrypt([]byte(str), key)
	}
	b.ReportAllocs()
}

func BenchmarkAesDecrypt(b *testing.B) {
	str := "我是努力的八王子ABCabc123!@#"
	key := []byte("12345678901234567890123456789012")
	crypted, _ := AesEncrypt([]byte(str), key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AesDecrypt(crypted, key)
	}
	b.ReportAllocs()
}