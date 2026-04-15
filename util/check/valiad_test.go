package check

import (
	"testing"
)

// go test -v valiad_test.go valiad.go

func TestFilteredSQLInject(t *testing.T) {
	// 应检测到注入
	str := "id OR (SELECT*FROM(SELECT(SLEEP(3)))plko) limit 1"
	if !IsSQLInject(str) {
		t.Error("should detect SQL injection:", str)
	}

	// 正常字符串不应被检测为注入
	normal := "hello world 123"
	if IsSQLInject(normal) {
		t.Error("should not detect SQL injection:", normal)
	}
}

func TestIsSQLInjectCompat(t *testing.T) {
	// 测试兼容函数
	str := "id OR (SELECT*FROM(SELECT(SLEEP(3)))plko) limit 1"
	if !IsSqlInject(str) {
		t.Error("IsSqlInject compat should detect SQL injection:", str)
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{2, true},
		{int64(100), true},
		{float64(3.14), true},
		{"aaa", false},
		{[]int{1, 2, 3, 4}, false},
		{"-123", true},
		{"-123.3", true},
		{"0xFF", true},
		{"+42", true},
		{"", false},
		{"  ", false},
		{"-", false},
		{"1.2.3", false},
		{"1e10", true},
		{"1.5e3", true},
	}
	for _, tt := range tests {
		result := IsNumeric(tt.input)
		if result != tt.expected {
			t.Errorf("IsNumeric(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsEmail(t *testing.T) {
	// 合法邮箱
	validEmails := []string{
		"test@example.com",
		"user1@mail.co.uk",
		"abc@test.org",
	}
	for _, email := range validEmails {
		if err := IsEmail(email); err != nil {
			t.Errorf("IsEmail(%s) should be valid, got err: %v", email, err)
		}
	}

	// 非法邮箱
	invalidEmails := []string{
		"test@",
		"@example.com",
		"testexample.com",
		"",
	}
	for _, email := range invalidEmails {
		if err := IsEmail(email); err == nil {
			t.Errorf("IsEmail(%s) should be invalid", email)
		}
	}
}

func TestIsMobile(t *testing.T) {
	// 合法手机号
	if !IsMobile("18675511217") {
		t.Error("IsMobile should accept 18675511217")
	}
	if !IsMobile("13800138000") {
		t.Error("IsMobile should accept 13800138000")
	}

	// 非法手机号
	if IsMobile("12345678901") {
		t.Error("IsMobile should reject 12345678901")
	}
	if IsMobile("1867551121") {
		t.Error("IsMobile should reject short number")
	}
	if IsMobile("abcdefghijk") {
		t.Error("IsMobile should reject non-numeric")
	}
}

func TestIsValidIDCardNumber(t *testing.T) {
	// 合法身份证号
	if !IsValidIDCardNumber("11010119900307617X") {
		t.Error("IsValidIDCardNumber should accept 11010119900307617X")
	}

	// 非法身份证号
	invalidIDs := []string{
		"123456789012345678",  // 格式不对
		"1101011990030761",   // 长度不够
		"11010119900307617A", // 最后一位非法
		"",                   // 空字符串
	}
	for _, id := range invalidIDs {
		if IsValidIDCardNumber(id) {
			t.Errorf("IsValidIDCardNumber should reject %s", id)
		}
	}
}

func TestIsValidIDCardCheckSum(t *testing.T) {
	// 合法身份证号（校验码正确）
	if !IsValidIDCardCheckSum("44138119880318213X") {
		t.Error("IsValidIDCardCheckSum should accept 44138119880318213X")
	}

	// 校验码错误的身份证号
	if IsValidIDCardCheckSum("440106199001011230") {
		t.Error("IsValidIDCardCheckSum should reject invalid checksum")
	}

	// 长度不对
	if IsValidIDCardCheckSum("12345") {
		t.Error("IsValidIDCardCheckSum should reject short input")
	}
}

func TestIsValidField(t *testing.T) {
	// 合法字段
	validFields := []string{
		"我.爱.宝贝kerr123-",
		"hello_world",
		"中文字段",
		"-0.2",
	}
	for _, f := range validFields {
		if err := IsValidField(f); err != nil {
			t.Errorf("IsValidField(%s) should be valid, got err: %v", f, err)
		}
	}

	// 非法字段
	invalidFields := []string{
		"",
		"abc!@#$%^&*",
		"=)(",
	}
	for _, f := range invalidFields {
		if err := IsValidField(f); err == nil {
			t.Errorf("IsValidField(%s) should be invalid", f)
		}
	}
}

func TestIsValidFields(t *testing.T) {
	// 全部合法
	if err := IsValidFields("我爱我家", "-0.2"); err != nil {
		t.Error("IsValidFields should pass for valid fields, got:", err)
	}

	// 包含非法字段
	err := IsValidFields("我爱我家", "abc!@#$%^&*", "-0.2", "=()")
	if err == nil {
		t.Error("IsValidFields should fail for invalid fields")
	} else {
		t.Log("expected error:", err)
	}

	// 空参数
	if err := IsValidFields(); err == nil {
		t.Error("IsValidFields should fail for empty args")
	}
}

func TestIsIP(t *testing.T) {
	// 合法IPv4
	if !IsIP("127.0.0.1") {
		t.Error("IsIP should accept 127.0.0.1")
	}
	if !IsIP("192.168.1.1") {
		t.Error("IsIP should accept 192.168.1.1")
	}

	// 合法IPv6
	if !IsIP("2001:0db8:86a3:08d3:1319:8a2e:0370:7344") {
		t.Error("IsIP should accept valid IPv6")
	}
	if !IsIP("::1") {
		t.Error("IsIP should accept ::1")
	}

	// 非法IP
	if IsIP("999.999.999.999") {
		t.Error("IsIP should reject 999.999.999.999")
	}
	if IsIP("not_an_ip") {
		t.Error("IsIP should reject not_an_ip")
	}
	if IsIP("") {
		t.Error("IsIP should reject empty string")
	}
}

func TestIsQQNumber(t *testing.T) {
	// 合法QQ号
	if err := IsQQNumber("414141"); err != nil {
		t.Error("IsQQNumber should accept 414141, got:", err)
	}
	if err := IsQQNumber("10001"); err != nil {
		t.Error("IsQQNumber should accept 10001, got:", err)
	}

	// 非法QQ号
	invalidQQs := []string{
		"1234",          // 太短
		"0123456",       // 0开头
		"12345abc",      // 含字母
		"",              // 空
		"123456789012345", // 太长（超过14位）
	}
	for _, qq := range invalidQQs {
		if err := IsQQNumber(qq); err == nil {
			t.Errorf("IsQQNumber should reject %q", qq)
		}
	}
}

func TestIsOpenid(t *testing.T) {
	// 合法openid
	if err := IsOpenid("wx_a123456789012345678901234567890"); err != nil {
		t.Error("IsOpenid should accept valid openid, got:", err)
	}

	// 非法openid
	invalidOpenids := []string{
		"short",                // 太短
		"",                     // 空
		"abc!@#$%^&*()1234567890123456789", // 含特殊字符
	}
	for _, oid := range invalidOpenids {
		if err := IsOpenid(oid); err == nil {
			t.Errorf("IsOpenid should reject %q", oid)
		}
	}
}

func TestHashString(t *testing.T) {
	dig := HashString("123")
	expected := "40bd001563085fc35165329ea1ff5c5ecbdbbeef"
	if dig != expected {
		t.Errorf("HashString(\"123\") = %s, want %s", dig, expected)
	}

	// 空字符串
	dig2 := HashString("")
	if dig2 == "" {
		t.Error("HashString(\"\") should not return empty string")
	}
}

func TestInNumRange(t *testing.T) {
	tests := []struct {
		val      interface{}
		min, max float64
		expected bool
	}{
		{"1.2", 0, 19, true},
		{5, 1, 10, true},
		{0, 0, 10, true},    // 边界：等于min
		{10, 0, 10, true},   // 边界：等于max
		{-1, 0, 10, false},  // 小于min
		{11, 0, 10, false},  // 大于max
		{"abc", 0, 10, false}, // 非数字
		{float64(3.14), 3.0, 4.0, true},
	}
	for _, tt := range tests {
		result := InNumRange(tt.val, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("InNumRange(%v, %v, %v) = %v, want %v", tt.val, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestIsSQLField(t *testing.T) {
	// 合法SQL字段名
	validFields := []string{
		"a1b",
		"user_name",
		"table.column",
		"my-field",
	}
	for _, f := range validFields {
		if !IsSQLField(f) {
			t.Errorf("IsSQLField(%s) should be true", f)
		}
	}

	// 非法SQL字段名
	invalidFields := []string{
		"abc!@#",
		"select *",
		"中文字段",
		"",
	}
	for _, f := range invalidFields {
		if IsSQLField(f) {
			t.Errorf("IsSQLField(%s) should be false", f)
		}
	}

	// 测试兼容函数
	if !IsSqlField("a1b") {
		t.Error("IsSqlField compat should accept a1b")
	}
}