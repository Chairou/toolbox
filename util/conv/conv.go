package conv

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	TIME_DATE      string = "2006-01-02"
	TIME_NORMAL    string = "2006-01-02 15:04:05"
	TIME_SHORTDATE string = "20060102"
)

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func String(val interface{}) string {
	if val == nil {
		return ""
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return ""
		}
		val = reValue.Interface()
		if val == nil {
			return ""
		}
		reValue = reflect.ValueOf(val)
	}
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func Int64(val interface{}) (int64, bool) {
	if val == nil {
		return 0, false
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return 0, false
		}
		val = reValue.Interface()
		if val == nil {
			return 0, false
		}
		reValue = reflect.ValueOf(val)
	}
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case []byte:
		return Int64(string(v))
	case string:
		v = strings.SplitN(v, ".", 2)[0]
		t, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return t, true
		} else {
			return 0, false
		}
	default:
		return 0, false
	}
}

func Uint64(val interface{}) (uint64, bool) {
	if val == nil {
		return 0, false
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return 0, false
		}
		val = reValue.Interface()
		if val == nil {
			return 0, false
		}
		reValue = reflect.ValueOf(val)
	}
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case int8:
		return uint64(v), true
	case int16:
		return uint64(v), true
	case int:
		return uint64(v), true
	case int32:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case float32:
		return uint64(v), true
	case float64:
		return uint64(v), true
	case []byte:
		return Uint64(string(v))
	case string:
		v = strings.SplitN(v, ".", 2)[0]
		t, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			return t, true
		} else {
			return 0, false
		}
	default:
		return 0, false
	}
}

func Int(val interface{}) (int, bool) {
	tmpVal, suc := Int64(val)
	return int(tmpVal), suc
}

func Uint(val interface{}) (uint, bool) {
	tmpVal, suc := Uint64(val)
	return uint(tmpVal), suc
}

func Float64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return 0, false
		}
		val = reValue.Interface()
		if val == nil {
			return 0, false
		}
		reValue = reflect.ValueOf(val)
	}
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case []byte:
		return Float64(string(v))
	case string:
		t, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return t, true
		} else {
			return 0, false
		}
	default:
		return 0, false
	}
}

func Bool(val interface{}) (bool, bool) {
	tmpVal, suc := Int64(val)
	return tmpVal != 0, suc
}

func IsNil(val interface{}) bool {
	if val == nil {
		return true
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() || reValue.IsNil() {
			return true
		}
		reValue = reflect.ValueOf(reValue.Interface())
	}
	return false
}

func Time(val interface{}) (time.Time, bool) {
	if val == nil {
		return time.Time{}, false
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return time.Time{}, false
		}
		val = reValue.Interface()
		if val == nil {
			return time.Time{}, false
		}
		reValue = reflect.ValueOf(val)
	}
	if val == nil {
		return time.Time{}, false
	}

	if v, ok := val.(time.Time); ok {
		return v, ok
	} else if v, ok := val.(string); ok {
		tlen := len(v)
		var t time.Time
		var err error
		switch tlen {
		case 8:
			t, err = time.ParseInLocation(TIME_SHORTDATE, v, time.Local)
		case 10:
			t, err = time.ParseInLocation(TIME_DATE, v, time.Local)
		case 19:
			t, err = time.ParseInLocation(TIME_NORMAL, v, time.Local)
		default:
			return t, false
		}
		if err != nil {
			return t, false
		} else {
			return t, true
		}
	} else {
		return time.Time{}, false
	}
}

func TimePtr(val interface{}) *time.Time {
	t, ok := Time(val)
	if ok {
		return &t
	} else {
		return nil
	}
}

// StringToArray ???shell???????????????Array
// "-a 123 -b hello" ---> ["-a","123","-b","hello"]
func StringToArray(ext string) (array []string, err error) {
	if len(ext) <= 0 {
		return array, nil
	}

	extRaw := []byte(strings.TrimSpace(ext))
	if len(extRaw) <= 0 {
		return array, nil
	}

	var tmp []byte
	//sOff ???????????????,doff ???????????????, escape \????????????
	var sOff, dOff, escape bool
	for offset := 0; offset < len(extRaw); offset++ {
		switch extRaw[offset] {
		case ' ':
			if sOff || dOff {
				tmp = append(tmp, extRaw[offset])
				continue
			}
			if tmp != nil {
				array = append(array, string(tmp))
				tmp = nil
			}
		case '"':
			//????????????????????????
			if dOff || escape {
				tmp = append(tmp, extRaw[offset])
				escape = false
				continue
			}
			//??????
			sOff = !sOff
			if !sOff && tmp == nil {
				//???????????????????????????????????????????????????????????????
				array = append(array, string(tmp))
			}
		case '\'':
			if sOff || escape {
				tmp = append(tmp, extRaw[offset])
				escape = false
				continue
			}
			dOff = !dOff
			if !dOff && tmp == nil {
				array = append(array, string(tmp))
			}
		case '\t':
			if sOff || dOff {
				tmp = append(tmp, extRaw[offset])
				continue
			}
		case '\\':
			//????????????
			if escape {
				tmp = append(tmp, extRaw[offset])
				escape = !escape
				continue
			}
			escape = !escape //true
		case '\r':
			fallthrough
		case '\n':
			if sOff || dOff {
				tmp = append(tmp, extRaw[offset])
				continue
			}
			if escape {
				//?????????????????????
				escape = !escape
				continue
			}
			if tmp != nil {
				array = append(array, string(tmp))
				tmp = nil
			}
		default:
			if escape {
				//????????????
				tmp = append(tmp, extRaw[offset-1])
				escape = !escape
			}
			tmp = append(tmp, extRaw[offset])
		}
		if offset == (len(extRaw)-1) && tmp != nil {
			array = append(array, string(tmp))
		}
	}

	if len(array)%2 != 0 {
		return nil, fmt.Errorf("array length is not even(%d), check your parameters", len(array))
	}

	return array, nil
}

// StringToMap ???shell???????????????Map
// "-a 123 -b hello" ---> {"-a":"123","-b":"hello"}
func StringToMap(ext string) (map[string]string, error) {
	array, err := StringToArray(ext)
	if len(array) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(array)/2)
	for i := range array {
		if i%2 == 0 {
			k := array[i]
			m[k] = array[i+1]
		}
	}
	return m, nil
}

// JsonToArray ???shell?????????json???????????????Array
// "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> ["-a","123","-b","hello"]
func JsonToArray(ext string) ([]string, error) {
	var array []string
	if len(ext) == 0 {
		return array, nil
	}
	extRaw := []byte(strings.TrimSpace(ext))
	extMap := make(map[string]string)
	err := jsoniter.Unmarshal(extRaw, &extMap)
	if err != nil {
		return array, err
	}
	keyCount := 0
	for key, value := range extMap {
		if strings.HasPrefix(value, "-") {
			continue
		}
		if strings.HasPrefix(key, "-") {
			array = append(array, key, value)
			keyCount++
		} else {
			array = append(array, "-"+key, value)
			keyCount++
		}
	}

	if keyCount != len(extMap) {
		return nil, fmt.Errorf("array length is not even(%d), check your parameters", len(array))
	}

	return array, nil
}

// JsonToString ???shell?????????json???????????????string
// "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> "-a 123 -b hello"
func JsonToString(ext string) (string, error) {
	var str string
	if len(ext) <= 0 {
		return str, nil
	}
	extRaw := []byte(strings.TrimSpace(ext))
	extMap := make(map[string]string)
	err := jsoniter.Unmarshal(extRaw, &extMap)
	if err != nil {
		return str, err
	}
	keyCount := 0
	for key, value := range extMap {
		if strings.HasPrefix(value, "-") {
			continue
		}
		if strings.HasPrefix(key, "-") {
			str += fmt.Sprintf("%s %s ", key, value)
			keyCount++
		} else {
			str += fmt.Sprintf("-%s %s ", key, value)
			keyCount++
		}
	}

	if keyCount != len(extMap) {
		return "", fmt.Errorf("json length is not even(%d), check your parameters", keyCount)
	}

	return strings.TrimSpace(str), nil
}

// StructToMap ???????????????Map[string]interface{} chairou??????
func StructToMap(in interface{}, tagName string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct { // ??????????????????????????????
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
	}

	t := v.Type()
	// ?????????????????????
	// ??????tagName??????map???key;????????????map???value
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		if tagValue := fi.Tag.Get(tagName); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}
