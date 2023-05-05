package check

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/conv"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// HashString hash字符串
func HashString(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// IsSQLInject 正则过滤sql注入的方法
func IsSQLInject(toMatchStr string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(?i)(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	return re.MatchString(toMatchStr)
}

// IsNumeric 验证数字类型
func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		return true
	case string:
		str := val.(string)
		if len(str) == 0 {
			return false
		}

		str = strings.Trim(str, " \t\r\n\v\f")
		if len(str) == 0 {
			return false
		}

		if str[0] == '-' || str[0] == '+' {
			if len(str) == 1 {
				return false
			}
			str = str[1:]
		}

		if len(str) > 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
			for _, h := range str[2:] {
				if !((h >= '0' && h <= '9') || (h >= 'a' && h <= 'f') || (h >= 'A' && h <= 'F')) {
					return false
				}
			}
			return true
		}
		// 0-9
		p, s, l := 0, 0, len(str)
		for i, v := range str {
			if v == '.' {
				if p > 0 || s > 0 || i+1 == l {
					return false
				}
				p = i
			} else if v == 'e' || v == 'E' {
				if i == 0 || s > 0 || i+1 == l {
					return false
				}
				s = i
			} else if v < '0' || v > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// SaltMD5 md5 hash
func SaltMD5(str string) (md5str string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("*T#e@c&h*%d*T#e@b&h*%s*T#e@a&h*%d", 0611, str, 1160)))) //将[]byte转成16进制
}

// IntToByte 整型转字节数组
func IntToByte(data int, len uintptr) (ret []byte) {
	ret = make([]byte, len)
	var tmp = 0xff
	var index uint
	for index = 0; index < uint(len); index++ {
		ret[index] = byte((tmp << (index * 8) & data) >> (index * 8))
	}
	return ret
}

func IsQQNumber(qq string) (err error) {
	pattern := "^[1-9][0-9]{4,13}$"
	matched, err := regexp.MatchString(pattern, qq)
	if err != nil {
		return err
	}
	if !matched {
		//长度验证
		if len(qq) < 5 {
			return fmt.Errorf("输入的QQ号[%s]格式不正确,请输入正确的QQ", qq)
		}
	}
	return nil
}

func IsOpenid(openid string) (err error) {
	//验证是否为openid
	pattern := "^[a-zA-Z0-9_-]{28,34}$"
	matched, err := regexp.MatchString(pattern, openid)
	if err != nil {
		return err
	}
	if !matched {
		//长度验证
		if len(openid) < 5 {
			return fmt.Errorf("输入的openid[%s]格式不正确,请输入正确的openid", openid)
		}
	}
	return nil
}

func IsEmail(email string) (err error) {
	//验证是否为email
	pattern := "^[a-z0-9]{1}[a-z0-9_-]{1,}@[a-z0-9]{1,}(.[a-z]{2,})*.[a-z]{2,}$"
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("输入的邮箱地址[%s]格式不正确,请输入正确的邮箱地址", email)
	}
	return nil
}

func IsMobile(mobile string) bool {
	reg := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(mobile)
}

// IsValidIDCardNumber 只能用到2099年, 到达2100年就会出错
func IsValidIDCardNumber(id string) bool {
	// 判断身份证号长度是否正确
	if len(id) != 18 {
		return false
	}

	// 判断身份证号前17位是否全是数字
	_, err := strconv.Atoi(id[:17])
	if err != nil {
		return false
	}

	// 判断身份证号的格式是否正确
	pattern := `^[\d]{6}(19|20)[\d]{2}(0[1-9]|1[0-2])(0[1-9]|[12][\d]|3[01])[\d]{3}[\dX]$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(id) {
		return false
	}

	return true
}

// IsValidIDCardCheckSum 增加了尾号校验
func IsValidIDCardCheckSum(idCard string) bool {
	// 长度检查
	if len(idCard) != 18 {
		return false
	}

	// 正则表达式匹配格式
	match, _ := regexp.MatchString(`^(\d{6})(\d{4})(\d{2})(\d{2})(\d{3})(\d|X)$`, idCard)
	if !match {
		return false
	}

	// 身份证校验码计算
	// 权重因子
	factors := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// 校验码
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	sum := 0
	for i := 0; i < len(factors); i++ {
		num, _ := strconv.Atoi(string(idCard[i]))
		sum += num * factors[i]
	}

	remainder := sum % 11
	checkCode := checkCodes[remainder]

	return checkCode == string(idCard[17])
}

// IsValidField 检查合法输入, 白名单, 汉字, 数字, 字母,下划线,点
func IsValidField(field string) (err error) {
	if len(field) <= 0 {
		return fmt.Errorf("field is null")
	}
	//验证是否为field
	pattern := `^[\p{Han}\p{Latin}0-9_,\.\-]+$`
	matched, err := regexp.MatchString(pattern, field)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("所传字段[%s]存在注入风险", field)
	}
	return nil
}

func IsValidFields(fields ...string) (err error) {
	if len(fields) <= 0 {
		err := errors.New("IsValidFields: field is null")
		return err
	}
	finalErrorBuf := strings.Builder{}
	finalErrorBuf.WriteString("IsValidFields:{ ")
	for _, arg := range fields {
		err := IsValidField(arg)
		if err != nil {
			finalErrorBuf.WriteString("【 ")
			finalErrorBuf.WriteString(arg)
			finalErrorBuf.WriteString(" is not Valid 】, ")
		}
	}
	finaStr := strings.TrimRight(finalErrorBuf.String(), ", ")

	finaStr += " }"
	if len(finalErrorBuf.String()) == len("IsValidFields:{ ") {
		return nil
	} else {
		return errors.New(finaStr)
	}
}

func IsIP(ip string) bool {
	address := net.ParseIP(ip)
	if address == nil {
		fmt.Println("ip地址格式不正确")
		return false
	} else {
		fmt.Println("正确的ip地址", address.String())
		return true
	}
}

func InNumRange(val interface{}, min, max float64) bool {
	mid, ok := conv.Float64(val)
	if !ok {
		return false
	}
	if math.Max(mid, min) == mid && math.Max(mid, max) == max {
		return true
	} else {
		return false
	}
}
