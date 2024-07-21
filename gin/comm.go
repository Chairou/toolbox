package gin

import (
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/conf"
	"github.com/Chairou/toolbox/logger"
	"github.com/Chairou/toolbox/util/check"
	"github.com/Chairou/toolbox/util/conv"
	"github.com/Chairou/toolbox/util/listopt"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"strings"
)

var log *logger.LogPool
var conf1 *conf.Config

func init() {
	var err error
	conf1 = conf.GetConf()
	log, err = logger.NewLogPool("api", conf1.LogFileName)
	if err != nil {
		_ = fmt.Errorf("NewLogPool err: %v", err)
		os.Exit(1)
	}
}

// Context wrap gin Context
type Context struct {
	*gin.Context
	requestID   string
	LoginMethod string
	UserName    string
}

// ParamConstruct ParamConstruct值类型
type ParamConstruct struct {
	FieldName    string
	DefaultValue interface{}
	CheckValue   []interface{}
	Need         bool
	Link         string
	Symbol       string
}

// ParamConByBlockstruct ParamConstruct值类型
type ParamConByBlockstruct struct {
	IsBlock     bool
	Link        string
	ParamConMap map[string]*ParamConstruct
}

// Debugf formats message according to format specifier
// and writes to log with level = Debug.
func (c *Context) Debugf(format string, params ...interface{}) {
	msg := fmt.Sprintf(c.requestID+" "+format, params...)
	log.Error(msg)
}

// Infof formats message according to format specifier
// and writes to log with level = Info.
func (c *Context) Infof(format string, params ...interface{}) {
	msg := fmt.Sprintf(c.requestID+" "+format, params...)
	log.Info(msg)

}

// Errorf formats message according to format specifier
// and writes to log with level = Error.
func (c *Context) Errorf(format string, params ...interface{}) error {
	msg := fmt.Sprintf(format, params...)
	log.Error(msg)
	return errors.New(msg)
}

// Debug formats message using the default formats for its operands
// and writes to log with level = Debug
func (c *Context) Debug(v ...interface{}) {
	msg := c.requestID + " " + fmt.Sprint(v...)
	log.Debug(msg)
}

// Info formats message using the default formats for its operands
// and writes to log with level = Info
func (c *Context) Info(v ...interface{}) {
	msg := c.requestID + " " + fmt.Sprint(v...)
	log.Info(msg)
}

// Error formats message using the default formats for its operands
// and writes to log with level = Error
func (c *Context) Error(v ...interface{}) error {
	msg := fmt.Sprint(v...)
	log.Error(c.requestID + " " + msg)
	return errors.New(msg)
}

// GetPager 获取分页参数
func (c *Context) GetPager() (pageIndex, pageSize uint, code int, err error) {
	pIndex := c.Query("pageIndex")
	if len(pIndex) == 0 {
		pageIndex = 1 //默认查第一页
	} else {
		mPageIndex, err := strconv.Atoi(pIndex)
		if err != nil || mPageIndex <= 0 {
			return pageIndex, pageSize, 59998, c.Error("pageIndex is invalid, pageIndex：" + pIndex)
		}
		pageIndex = uint(mPageIndex)
	}
	pSize := c.Query("pageSize")
	if len(pSize) == 0 {
		pageSize = 10000 //默认查10000条
	} else {
		mPageSize, err := strconv.Atoi(pSize)
		if err != nil || mPageSize <= 0 {
			return pageIndex, pageSize, 59999, c.Error("pageSize is invalid, pageSize：" + pSize)
		}
		pageSize = uint(mPageSize)
	}
	return pageIndex, pageSize, code, nil
}

// GetConditionByParam 根据参数生成sql查询条件并检测前端参数正确性,支持无参无条件情况
func (c *Context) GetConditionByParam(parConstruct map[string]*ParamConstruct) (string, []interface{}, error) {
	if len(parConstruct) == 0 {
		return "", nil, nil
	}

	strCondition := " 1=1 "

	isNotAllCondition, ok := c.Get("isNotAllCondition")
	if ok && isNotAllCondition.(bool) == true {
		strCondition = " "
	}

	args := make([]interface{}, 0)
	orderByStr := ""
	for k, v := range parConstruct {
		strParam := c.Query(k)
		//1.无值且必传,直接返回
		if len(strParam) == 0 && v.Need {
			return "", nil, c.Error("need param " + k + " is null.")
		}
		//2.验证参数值
		if len(strParam) > 0 {
			timeStr := strings.TrimSpace(strParam)
			if len(timeStr) <= 10 {
				switch k {
				case "startTime", "starttime":
					strParam = timeStr + " 00:00:00"
				case "endTime", "endtime":
					strParam = timeStr + " 23:59:59"
				}
			}
			err := c.CheckParam(k, strParam, v.CheckValue)
			if err != nil {
				return "", nil, err
			}
		}
		//3.根据key生成where条件
		switch k {
		case "orderBy":
			//orderBy特殊处理
			orderByStr = GenOrder(strParam, v)
		case "searchKey":
			//searchKey特殊处
			strSearchKey, agsSearchKey := GenSearchKeyWhere(strParam, v.FieldName, v.Link, v.Symbol)
			if len(strSearchKey) > 0 && len(agsSearchKey) > 0 {
				strCondition += strSearchKey
				args = append(args, agsSearchKey...)
			}
		case "accessPerson":
			//accessPerson特殊处理,创建(接入)人查询,不传或传0表示查所有人;1表示本人创建;2表示同事创建
			strAccessPerson, agsAccessPerson := genAccessPersonWhere(c, strParam, v.Symbol, v.FieldName, v.Link)
			if len(strAccessPerson) > 0 && len(agsAccessPerson) > 0 {
				strCondition += strAccessPerson
				args = append(args, agsAccessPerson...)
			}
		default:
			//其它参数
			arg := genArgs(strParam, v)
			if arg != nil && len(conv.String(arg)) > 0 {
				strCondition += genCondition(v)
				args = append(args, arg)
			}
		}
	}
	if len(orderByStr) > 0 {
		strCondition += orderByStr
	}

	c.Info("GetConditionByParam strCondition:=======", strCondition, ";args:", args)

	return strCondition, args, nil
}

// CheckParam 检验参数值
func (c *Context) CheckParam(key string, param interface{}, defaultValue []interface{}) error {
	c.Info("CheckParam key:", key, ";param:", param, ";defaultValue:", defaultValue)
	//1.参数值为空,返回错误
	if param == nil || conv.String(param) == "" {
		return c.Error("param " + key + " is null.")
	}
	//2.没有定义取值范围数组,表示参数有值,正常返回nil
	if defaultValue == nil || len(defaultValue) == 0 {
		return nil
	}
	//3.有取值范围数组,判断该参数值是否在取值范围内
	var isIn = false
	for _, v := range defaultValue {
		if conv.String(v) == conv.String(param) {
			isIn = true
			break
		}
	}
	//在取值范围内,正常返回nil
	if isIn {
		return nil
	}
	//不在取值范围内,返回错误
	return c.Error("param " + key + " value is error.")
}

// genCondition 生成条件
func genCondition(paramConstruct *ParamConstruct) string {
	if strings.Contains(paramConstruct.FieldName, "|") {
		fields := strings.Split(paramConstruct.FieldName, "|")
		retCondition := " " + paramConstruct.Link + " ("
		for k, fieldName := range fields {
			oneCondition := genOneCondition(fieldName, "or", paramConstruct.Symbol)
			if k == 0 {
				oneCondition = strings.TrimPrefix(oneCondition, " or ")
			}
			retCondition += oneCondition
		}
		return retCondition + ") "
	}
	return genOneCondition(paramConstruct.FieldName, paramConstruct.Link, paramConstruct.Symbol)
}

func genOneCondition(fieldName, link, symbol string) string {
	if symbol == "FIND_IN_SET" {
		//" " + [and/or] + " " + "FIND_IN_SET(?," + [fieldName] + ") "
		return " " + link + " " + "FIND_IN_SET(?, " + fieldName + ") "
	}
	//" " + [and/or] + " " +  [fieldName] + " " + [=/<>/!=/>/</>=/<=/like/not like] + " ? "
	return " " + link + " " + fieldName + " " + symbol + " ? "
}

// GenOrder 生成OrderBy条件
func GenOrder(param string, paramConstruct *ParamConstruct) string {
	if len(param) == 0 {
		param = conv.String(paramConstruct.DefaultValue)
	}
	if len(param) == 0 {
		return ""
	}

	isExp := false
	orderArr := make([]string, 0)
	for _, group := range strings.Split(param, ";") {
		orderBy := strings.Split(group, "|")
		if orderBy != nil && len(orderBy[:]) == 2 && (orderBy[1] == "asc" || orderBy[1] == "desc") {
			field := orderBy[0]
			if ok := check.IsSqlField(field); ok == true {
				orderArr = append(orderArr, fmt.Sprintf("%s %s", field, orderBy[1]))
				isExp = true
			}
		}
	}

	if isExp {
		return fmt.Sprintf(" order by %s ", strings.Join(orderArr, ","))
	}

	return ""
}

// genArgs 生成条件值
func genArgs(param string, paramConstruct *ParamConstruct) interface{} {
	var value interface{}
	if len(param) == 0 {
		value = paramConstruct.DefaultValue
	} else {
		value = param
	}
	nValue := conv.String(value)
	if len(nValue) > 0 && (paramConstruct.Symbol == "like" || paramConstruct.Symbol == "not like") {
		return "%" + nValue + "%"
	}
	return value
}

// GenSearchKeyWhere 生成关键字查询的特殊条件及值
func GenSearchKeyWhere(param, fieldName, link, symbol string) (condition string, args []interface{}) {
	if len(param) == 0 {
		return condition, args
	}
	var conditionArr []string
	_, e := strconv.Atoi(param)
	fields := strings.Split(fieldName, "|")
	for _, fieldName := range fields {
		if len(symbol) > 0 {
			if symbol == "FIND_IN_SET" {
				conditionArr = append(conditionArr, fmt.Sprintf("FIND_IN_SET(?, %s)", fieldName))
				args = append(args, param)
			} else if listopt.IsInStringArr([]string{"=", "like", "not like"}, symbol) {
				conditionArr = append(conditionArr, fmt.Sprintf("%s %s ?", fieldName, symbol))
				if symbol == "=" { //数值型,用精确查询
					args = append(args, param)
				} else { //字符串,用模糊查询
					args = append(args, "%"+param+"%")
				}
			}
		} else {
			if e == nil { //数值型,用精确查询
				conditionArr = append(conditionArr, fieldName+" = ?")
				args = append(args, param)
			} else { //字符串,用模糊查询
				conditionArr = append(conditionArr, fieldName+" like ?")
				args = append(args, "%"+param+"%")
			}
		}
	}
	if len(conditionArr) > 0 {
		condition = fmt.Sprintf(" %s (%s) ", link, strings.Join(conditionArr, " or "))
	}
	return condition, args
}

// genArgs 生成按创建人查询的特殊条件及值
func genAccessPersonWhere(c *Context, param, symbol, fieldName, link string) (condition string, args []interface{}) {
	//param: "1" 、 "2"
	if param == "" || param == "0" || len(param) == 0 {
		return condition, args
	}
	//Symbol: "=|!=" 、 "=|<>" 、 "like|not like"
	symbols := strings.Split(symbol, "|")
	if len(symbols) != 2 {
		return condition, args
	}
	value := c.UserName
	if len(value) <= 0 {
		return condition, args
	}
	var nSymbol string
	if param == "1" { //self
		nSymbol = symbols[0]
	} else if param == "2" { //colleague
		nSymbol = symbols[1]
	} else { //all
		return condition, args
	}
	condition = genCondition(&ParamConstruct{
		FieldName: fieldName,
		Link:      link,
		Symbol:    nSymbol,
	})
	if nSymbol == "like" || nSymbol == "not like" {
		value = "%" + value + "%"
	}
	args = append(args, value)
	return condition, args
}

// GetConditionByParamBlock 检测前端参数并根据参数生成sql查询条件,支持无参无条件情况
func (c *Context) GetConditionByParamBlock(parConByBlockstruct []*ParamConByBlockstruct) (string, []interface{}, error) {
	if len(parConByBlockstruct) == 0 {
		return "", nil, nil
	}

	strCondition := " 1=1 "

	args := make([]interface{}, 0)

	for _, pbb := range parConByBlockstruct {
		if pbb.IsBlock {
			strC, arg, err := c.GetBaseConditionByParam(pbb.ParamConMap)
			if err != nil {
				return "", nil, err
			}
			if len(strC) > 0 && len(arg) > 0 {
				strCondition += " " + pbb.Link + " (" + strC + ") "
				args = append(args, arg...)
			}
		} else {
			strC, arg, err := c.GetConditionByParam(pbb.ParamConMap)
			if err != nil {
				return "", nil, err
			}
			strCondition += " " + pbb.Link + strC
			args = append(args, arg...)
		}
	}

	return strCondition, args, nil
}

// GetBaseConditionByParam 检测前端参数并根据参数生成sql查询条件,只验证不关注无参无条件情况
func (c *Context) GetBaseConditionByParam(parConstruct map[string]*ParamConstruct) (string, []interface{}, error) {
	if len(parConstruct) == 0 {
		return "", nil, nil
	}

	strCondition := ""

	args := make([]interface{}, 0)
	for k, v := range parConstruct {
		strParam := c.Query(k)
		//1.无值且必传,直接返回
		if len(strParam) == 0 && v.Need {
			return "", nil, c.Error("need param " + k + " is null.")
		}
		//2.验证参数值
		if len(strParam) > 0 {
			switch k {
			case "startTime":
				strParam = strings.TrimSpace(strParam) + " 00:00:00"
			case "endTime":
				strParam = strings.TrimSpace(strParam) + " 23:59:59"
			}
			err := c.CheckParam(k, strParam, v.CheckValue)
			if err != nil {
				return "", nil, err
			}
		}
		//3.根据key生成where条件
		switch k {
		case "orderBy":
			//orderBy特殊处理
			return "", nil, c.Error("orderBy not allowed here.")
		case "searchKey":
			//searchKey特殊处理,模糊查找无值不需要执行like
			strSearchKey, agsSearchKey := GenBaseSearchKeyWhere(strParam, v.FieldName, v.Symbol)
			if len(strSearchKey) > 0 && len(agsSearchKey) > 0 {
				if len(strCondition) > 0 {
					strSearchKey = " " + v.Link + strSearchKey
				}
				strCondition += strSearchKey
				args = append(args, agsSearchKey...)
			}
		case "accessPerson":
			return "", nil, c.Error("accessPerson not allowed here.")
		default:
			//其它参数
			arg := genArgs(strParam, v)
			if arg != nil && len(conv.String(arg)) > 0 {
				strOther := " " + v.FieldName + " " + v.Symbol + " ? "
				if len(strCondition) > 0 {
					strOther = " " + v.Link + strOther
				}
				strCondition += strOther
				args = append(args, arg)
			}
		}
	}

	c.Info("GetBaseConditionByParam:", strCondition)

	return strCondition, args, nil
}

// GenBaseSearchKeyWhere 生成关键字查询的特殊条件及值
func GenBaseSearchKeyWhere(param, fieldName, symbol string) (condition string, args []interface{}) {
	if len(param) == 0 {
		return condition, args
	}
	var conditionArr []string
	_, e := strconv.Atoi(param)
	for _, fieldName := range strings.Split(fieldName, "|") {
		if len(symbol) > 0 && listopt.IsInStringArr([]string{"=", "like", "not like"}, symbol) {
			conditionArr = append(conditionArr, fmt.Sprintf("%s %s ?", fieldName, symbol))
			if symbol == "=" { //数值型,用精确查询
				args = append(args, param)
			} else { //字符串,用模糊查询
				args = append(args, "%"+param+"%")
			}
		} else {
			if e == nil { //数值型,用精确查询
				conditionArr = append(conditionArr, fieldName+" = ?")
				args = append(args, param)
			} else { //字符串,用模糊查询
				conditionArr = append(conditionArr, fieldName+" like ?")
				args = append(args, "%"+param+"%")
			}
		}
	}
	if len(conditionArr) > 0 {
		condition = fmt.Sprintf(" (%s) ", strings.Join(conditionArr, " or "))
	}
	return condition, args
}
