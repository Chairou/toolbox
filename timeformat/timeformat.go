package timeformat

import (
	"github.com/Chairou/toolbox/util/conv"
	jsoniter "github.com/json-iterator/go"
	"time"
	"unsafe"
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
