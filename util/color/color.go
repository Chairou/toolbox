package color

import (
	"github.com/Chairou/toolbox/util/conv"
	"regexp"
	"strings"
)

var (
	Reset = "\033[0m"

	/////////////
	// Special //
	/////////////

	Bold      = "\033[1m"
	Underline = "\033[4m"

	/////////////////
	// Text colors //
	/////////////////

	Black  = "\033[30m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"

	///////////////////////
	// Background colors //
	///////////////////////

	BlackBackground  = "\033[40m"
	RedBackground    = "\033[41m"
	GreenBackground  = "\033[42m"
	YellowBackground = "\033[43m"
	BlueBackground   = "\033[44m"
	PurpleBackground = "\033[45m"
	CyanBackground   = "\033[46m"
	GrayBackground   = "\033[47m"
	WhiteBackground  = "\033[107m"
)

func SetColor(color string, v interface{}) string {
	str := conv.String(v)
	if strings.HasPrefix(color, "\033[") {
		ret := strings.Builder{}
		ret.WriteString(color)
		ret.WriteString(str)
		ret.WriteString(Reset)
		return ret.String()
	} else {
		return str
	}
}

func SetFbColor(fontColor, backGroupColor string, v interface{}) string {
	str := conv.String(v)
	if strings.HasPrefix(fontColor, "\033[") && strings.HasPrefix(backGroupColor, "\033[") {
		ret := strings.Builder{}
		ret.WriteString(fontColor)
		ret.WriteString(backGroupColor)
		ret.WriteString(str)
		ret.WriteString(Reset)
		return ret.String()
	} else {
		return str
	}
}

func RemoveColor(str string) string {
	re := regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)
	cleanStr := re.ReplaceAllString(str, "")
	text := extractText(cleanStr)

	return text
}

func extractText(str string) string {
	var text string
	inEscape := false
	for _, ch := range str {
		if ch == '\033' {
			inEscape = true
		} else if inEscape && ch >= 'A' && ch <= 'Z' {
			inEscape = false
		} else if inEscape && ch >= 'a' && ch <= 'z' {
			inEscape = false
		} else if !inEscape {
			text += string(ch)
		}
	}
	return text
}
