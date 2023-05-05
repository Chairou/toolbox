package check

import (
	"testing"
)

//  go test -v valiad_test.go valiad.go

func TestFilteredSQLInject(t *testing.T) {
	str := "id OR (SELECT*FROM(SELECT(SLEEP(3)))plko) limit 1"
	if IsSQLInject(str) {
		t.Log("true, has injected")
	} else {
		t.Error("false, filter error")
	}
}

func TestIsNumeric(t *testing.T) {
	i := 2
	j := "aaa"
	k := []int{1, 2, 3, 4}
	l := "-123"
	m := "-123.3"
	if !IsNumeric(i) {
		t.Error("IsNumeric err:", i)
	}
	if IsNumeric(j) {
		t.Error("IsNumeric err:", j)
	}
	if IsNumeric(k) {
		t.Error("IsNumeric err:", k)
	}
	if !IsNumeric(l) {
		t.Error("IsNumeric err:", l)
	}
	if !IsNumeric(m) {
		t.Error("IsNumeric err:", m)
	}
}

func TestCheckEmail(t *testing.T) {
	email := "test@example.com"
	err := IsEmail(email)
	if err != nil {
		t.Error(email, err)
	}
}

func TestCheckMobile(t *testing.T) {
	mobileNumber := "18675511217"
	ok := IsMobile(mobileNumber)
	if ok == true {
		t.Log("IsMobile correct")
	} else {
		t.Error("IsMobile wrong")
	}
}

func TestCheckIdCard(t *testing.T) {
	idCard := "11010119900307617X"
	ok := IsValidIDCardNumber(idCard)
	if ok != true {
		t.Error("IsValidIDCardNumber wrong", idCard)
	}
}

// TestCheckId, 推荐用此函数IsValidIDCardCheckSum
func TestCheckId(t *testing.T) {
	idCard := "44138119880318213X"
	ok := IsValidIDCardCheckSum(idCard)
	if ok != true {
		t.Error("IsValidIDCardNumber wrong", idCard)
	}
}

func TestCheckField(t *testing.T) {
	str := "我.爱.宝贝kerr123-"
	err := IsValidField(str)
	if err != nil {
		t.Error("IsValidField err:", err)
	}
}

func TestCheckIP(t *testing.T) {
	ip := "127.0.0.1"
	ok := IsIP(ip)
	if ok == false {
		t.Error("IsIP err:", ip)
	}
	ip = "2001:0db8:86a3:08d3:1319:8a2e:0370:7344"
	ok = IsIP(ip)
	if ok == false {
		t.Error("IsIP err:", ip)
	}
}

func TestCheckQQNumber(t *testing.T) {
	qqNumber := "414141"
	err := IsQQNumber(qqNumber)
	if err != nil {
		t.Error("CheckQQNumber err:", err)
	}
}

func TestIntToByte(t *testing.T) {
	var aa int = 16
	byteList := IntToByte(aa, 4)
	t.Log(byteList)
	if byteList[0] != 16 {
		t.Error("CheckIntToByte err:")
	}
}

func TestSaltMD5(t *testing.T) {
	source := "123"
	dig := SaltMD5(source)
	if dig != "6226dd203e19b0bc02ee41af34275a44" {
		t.Error("CheckSaltMD5 err:", dig)
	}
}

func TestIsOpenid(t *testing.T) {
	openid := "wx_a123456789012345678901234567890"
	err := IsOpenid(openid)
	if err != nil {
		t.Error("CheckIsOpenid err:", err)
	}
}

func TestHashString(t *testing.T) {
	source := "123"
	dig := HashString(source)
	if dig != "40bd001563085fc35165329ea1ff5c5ecbdbbeef" {
		t.Error("HashString err:", dig)
	}
}

func TestIsValidFields(t *testing.T) {
	a1 := "我爱我家"
	b1 := "abc!@#$%^&*"
	c1 := "-0.2"
	d1 := "=)("
	err := IsValidFields(a1, c1)
	if err != nil {
		t.Error(err)
	}
	err = IsValidFields(a1, b1, c1, d1)
	if err != nil {
		t.Log(err)
	}

}
