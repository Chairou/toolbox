package check

import (
	"testing"
)

//  go test -v valiad_test.go checkValiad.go

func TestFilteredSQLInject(t *testing.T) {
	str := "id OR (SELECT*FROM(SELECT(SLEEP(3)))plko) limit 1"
	if FilteredSQLInject(str) {
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
	err := CheckEmail(email)
	if err != nil {
		t.Error(email, err)
	}
}

func TestCheckMobile(t *testing.T) {
	mobileNumber := "18675511217"
	ok := CheckMobile(mobileNumber)
	if ok == true {
		t.Log("CheckMobile correct")
	} else {
		t.Error("CheckMobile wrong")
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
	err := CheckField(str)
	if err != nil {
		t.Error("CheckField err:", err)
	}
}

func TestCheckIP(t *testing.T) {
	ip := "127.0.0.1"
	ok := CheckIP(ip)
	if ok == false {
		t.Error("CheckIP err:", ip)
	}
	ip = "2001:0db8:86a3:08d3:1319:8a2e:0370:7344"
	ok = CheckIP(ip)
	if ok == false {
		t.Error("CheckIP err:", ip)
	}
}
