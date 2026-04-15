// Package check 提供常用的数据校验工具函数
package check

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/Chairou/toolbox/util/conv"
)

// 预编译正则表达式，避免重复编译
var (
	sqlInjectRe  = regexp.MustCompile(`(?:')|(?:--)|(/\*(?:.|[\n\r])*?\*/)|(?i)(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`)
	qqNumberRe   = regexp.MustCompile(`^[1-9][0-9]{4,13}$`)
	openidRe     = regexp.MustCompile(`^[a-zA-Z0-9_-]{28,34}$`)
	emailRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]+@[a-z0-9]+(\.[a-z]{2,})*\.[a-z]{2,}$`)
	mobileRe     = regexp.MustCompile(`^1([38][0-9]|14[579]|5[^4]|16[6]|7[0-35-8]|9[189])\d{8}$`)
	idCardRe     = regexp.MustCompile(`^[\d]{6}(19|20)[\d]{2}(0[1-9]|1[0-2])(0[1-9]|[12][\d]|3[01])[\d]{3}[\dX]$`)
	idCardSumRe  = regexp.MustCompile(`^(\d{6})(\d{4})(\d{2})(\d{2})(\d{3})(\d|X)$`)
	validFieldRe = regexp.MustCompile(`^[\p{Han}\p{Latin}0-9_,\.\-]+$`)
	sqlFieldRe   = regexp.MustCompile(`^[\p{Latin}0-9_\.\-]+$`)
)

// HashString 对字符串进行SHA1哈希
func HashString(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// IsSQLInject 正则过滤sql注入的方法
func IsSQLInject(toMatchStr string) bool {
	return sqlInjectRe.MatchString(toMatchStr)
}

// IsSqlInject 保留旧函数名以保持向后兼容
// Deprecated: 请使用 IsSQLInject
func IsSqlInject(toMatchStr string) bool {
	return IsSQLInject(toMatchStr)
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
	default:
		return false
	}
}

// IsQQNumber 验证QQ号格式是否正确
func IsQQNumber(qq string) error {
	if !qqNumberRe.MatchString(qq) {
		return fmt.Errorf("输入的QQ号[%s]格式不正确,请输入正确的QQ", qq)
	}
	return nil
}

// IsOpenid 验证openid格式是否正确
func IsOpenid(openid string) error {
	if !openidRe.MatchString(openid) {
		return fmt.Errorf("输入的openid[%s]格式不正确,请输入正确的openid", openid)
	}
	return nil
}

// IsEmail 验证邮箱地址格式是否正确
func IsEmail(email string) error {
	if !emailRe.MatchString(email) {
		return fmt.Errorf("输入的邮箱地址[%s]格式不正确,请输入正确的邮箱地址", email)
	}
	return nil
}

// IsMobile 验证手机号格式是否正确
func IsMobile(mobile string) bool {
	return mobileRe.MatchString(mobile)
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
	if !idCardRe.MatchString(id) {
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
	if !idCardSumRe.MatchString(idCard) {
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

// IsValidField 检查合法输入, 白名单, 汉字, 数字, 字母, 下划线, 点
func IsValidField(field string) error {
	if len(field) <= 0 {
		return fmt.Errorf("field is null")
	}
	if !validFieldRe.MatchString(field) {
		return fmt.Errorf("所传字段[%s]存在注入风险", field)
	}
	return nil
}

// IsValidFields 批量检查多个字段的合法性
func IsValidFields(fields ...string) error {
	if len(fields) <= 0 {
		return errors.New("IsValidFields: field is null")
	}
	finalErrorBuf := strings.Builder{}
	finalErrorBuf.WriteString("IsValidFields:{ ")
	hasError := false
	for _, arg := range fields {
		err := IsValidField(arg)
		if err != nil {
			hasError = true
			finalErrorBuf.WriteString("【 ")
			finalErrorBuf.WriteString(arg)
			finalErrorBuf.WriteString(" is not Valid 】, ")
		}
	}
	if !hasError {
		return nil
	}
	finalStr := strings.TrimRight(finalErrorBuf.String(), ", ")
	finalStr += " }"
	return errors.New(finalStr)
}

// IsIP 验证IP地址格式是否正确
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// InNumRange 检查数值是否在[min, max]范围内
func InNumRange(val interface{}, min, max float64) bool {
	mid, ok := conv.Float64(val)
	if !ok {
		return false
	}
	return mid >= min && mid <= max
}

// IsSQLField 检查字符串是否为合法的SQL字段名
func IsSQLField(str string) bool {
	return sqlFieldRe.MatchString(str)
}

// IsSqlField 保留旧函数名以保持向后兼容
// Deprecated: 请使用 IsSQLField
func IsSqlField(str string) bool {
	return IsSQLField(str)
}

// 以下保留旧函数签名的兼容别名（返回值带命名的旧版本）
// 注意：旧版 IsQQNumber/IsOpenid/IsEmail/IsValidField/IsValidFields 的返回值签名
// 从 (err error) 改为 error，这是向后兼容的，调用方无需修改