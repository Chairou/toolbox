package timeformat

import (
	"database/sql/driver"
	"fmt"
	"time"
	"unsafe"

	"github.com/Chairou/toolbox/util/conv"
	jsoniter "github.com/json-iterator/go"
)

const (
	DefaultFormat      = "2006-01-02 15:04:05"
	DateFormart        = "2006-01-02"
	TimeFormart        = "15:04:05"
	MicordSecondFormat = "2006-01-02 15:04:05.99999"
)

func init() {
	jsoniter.RegisterTypeEncoder("time.Time", &TimeFormat{})
	jsoniter.RegisterTypeDecoder("time.Time", &TimeFormat{})
}

type TimeFormat struct{}

func NowString() string {
	return time.Now().Format(DefaultFormat)
}

func (codec *TimeFormat) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	*((*time.Time)(ptr)), _ = conv.Time(iter.ReadString())
}
func (codec *TimeFormat) IsEmpty(ptr unsafe.Pointer) bool {
	ts := *((*time.Time)(ptr))
	return ts.UnixNano() == 0
}
func (codec *TimeFormat) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(DefaultFormat))
}

func (codec *TimeFormat) Date(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(DateFormart))
}

func (codec *TimeFormat) Time(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(TimeFormart))
}

type Time time.Time

const (
	datetime = "2006-01-02 15:04:05"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+datetime+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(datetime)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, datetime)
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	return time.Time(t).Format(datetime)
}

// Value 实现 driver.Valuer 接口，写入数据库时将 timeformat.Time 转为标准 time.Time
func (t Time) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取时将值转为 timeformat.Time
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		*t = Time(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = Time(v)
		return nil
	case []byte:
		// MySQL 有时会返回 []byte 格式的时间字符串
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", string(v), time.Local)
		if err != nil {
			return fmt.Errorf("timeformat.Time Scan: cannot parse %q: %w", string(v), err)
		}
		*t = Time(parsed)
		return nil
	case string:
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			return fmt.Errorf("timeformat.Time Scan: cannot parse %q: %w", v, err)
		}
		*t = Time(parsed)
		return nil
	default:
		return fmt.Errorf("timeformat.Time Scan: unsupported type %T", value)
	}
}
