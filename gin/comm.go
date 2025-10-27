package gin

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/conf"
	"os"

	"github.com/Chairou/toolbox/logger"
	"github.com/Chairou/toolbox/util/check"
	"github.com/Chairou/toolbox/util/conv"
	"github.com/Chairou/toolbox/util/listopt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const API_OK = 0
const API_INTERNAL_ERROR = -99
const API_DB_ERROR = -98
const API_REMOTE_ERROR = -97
const API_ARG_ERROR = -96

var log *logger.LogPool
var conf1 *conf.Config

type H map[string]any

const _RequestIDKey = "__RequestID__"
const _ContextKey = "__XContext__"

// Context wrap gin Context
type Context struct {
	*gin.Context
	requestID   string
	LoginMethod string
	UserName    string
}

type Ret struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
	Seq  string      `json:"seq"`
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

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

type StdHandlerFunc func(*Context)

type responseDumper struct {
	gin.ResponseWriter
	outBuffer *bytes.Buffer
}

func NewServer(env string, logFileName string, middle []func(c *Context)) *gin.Engine {
	var err error
	log, err = logger.NewLogPool("api", logFileName)
	if err != nil {
		_ = fmt.Errorf("NewLogPool err: %v", err)
		os.Exit(1)
	}
	r := gin.Default()

	mode := env
	switch mode {
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	stdRouter := &RouterGroup{
		routerGroup: &r.RouterGroup,
	}
	stdRouter.Use(SafeCheck)
	stdRouter.Use(ResponseRecorder)

	for _, v := range middle {
		stdRouter.Use(v)
	}

	SetupRouter(stdRouter)

	return r
}

var routerRegisters []func(*RouterGroup)

// SetRouterRegister 设置路由注册器
func SetRouterRegister(reg func(group *RouterGroup)) {
	routerRegisters = append(routerRegisters, reg)
}

// SetupRouter 设置路由
func SetupRouter(group *RouterGroup) {
	for _, reg := range routerRegisters {
		reg(group)
	}
}

func (r *responseDumper) Write(data []byte) (int, error) {
	r.outBuffer.Write(data)
	return r.ResponseWriter.Write(data)
}

func (r *responseDumper) bytes() []byte {
	if r.ResponseWriter.Header().Get("Content-Encoding") == "gzip" {
		return []byte("[gzip data]")
	}
	return r.outBuffer.Bytes()
}

// RetJson 直接返回json串
func (c *Context) RetJson(code int, data interface{}, messages ...interface{}) {
	var ret Ret
	ret.Code = code
	ret.Data = data
	seq, ok := c.Get("seq")
	if ok {
		ret.Seq = conv.String(seq)
	} else {
		ret.Seq = ""
	}

	msg := strings.Builder{}
	for _, arg := range messages {
		switch v := arg.(type) {
		case error:
			msg.WriteString(v.Error())
		case string:
			// 处理string类型
			msg.WriteString(v)
		case int:
			// 处理int类型
			msg.WriteString(fmt.Sprintf("%d", v))
		case int32:
			// 处理int类型
			msg.WriteString(fmt.Sprintf("%d", v))
		case int64:
			// 处理int类型
			msg.WriteString(fmt.Sprintf("%d", v))
		case float32:
			// 处理float64类型
			msg.WriteString(fmt.Sprintf("%f", v))
		case float64:
			// 处理float64类型
			msg.WriteString(fmt.Sprintf("%f", v))
		case bool:
			// 处理bool类型
			msg.WriteString(fmt.Sprintf(" %t", v))
		default:
			// 处理其他类型
			msg.WriteString(fmt.Sprintf(" %v", v))
		}
	}
	ret.Msg = msg.String()
	fmt.Println(ret)
	c.JSON(http.StatusOK, ret)

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

// GetLimit 获取分页参数:PageIndex,PageSize，返回offset,limit
func (c *Context) GetLimit() (offset, limit int, err error) {
	pageIndex, pageSize, _, err := c.GetPager()
	if err != nil {
		return 0, 0, err
	}
	offset = (int(pageIndex) - 1) * int(pageSize)
	limit = int(pageSize)
	return offset, limit, nil
}

// GetConditionByParam 根据参数生成sql查询条件并检测前端参数正确性,支持无参无条件情况
func (c *Context) GetConditionByParam(parConstruct map[string]*ParamConstruct) (string, []interface{}, string, error) {
	if len(parConstruct) == 0 {
		return "", nil, "", nil
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
			return "", nil, "", c.Error("need param " + k + " is null.")
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
				return "", nil, "", err
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

	return strCondition, args, orderByStr, nil
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

	//if isExp {
	//	return fmt.Sprintf(" order by %s ", strings.Join(orderArr, ","))
	// }
	// 为了适应 gorm，需要去掉 order by
	if isExp {
		return fmt.Sprintf(" %s ", strings.Join(orderArr, ","))
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
			strC, arg, orderStr, err := c.GetConditionByParam(pbb.ParamConMap)
			if err != nil {
				return "", nil, err
			}
			strCondition += " " + pbb.Link + strC + orderStr
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

type HandlerFunc func(*Context)

// RouterGroup wrap gin RouterGroup
type RouterGroup struct {
	routerGroup *gin.RouterGroup
}

func (r *RouterGroup) SetGroup(rg *gin.RouterGroup) {
	r.routerGroup = rg
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func wrapHandlers(handlers []HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		ginHandlers[i] = wrapHandler(h)
	}
	return ginHandlers
}

func wrapHandler(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {

		requestID := c.GetString(_RequestIDKey)
		if len(requestID) == 0 {
			if requestID = c.Request.Header.Get("X-Request-Id"); len(requestID) > 0 {
				c.Set(_RequestIDKey, requestID)
				c.Writer.Header().Set("X-Request-Id", requestID)
			} else {
				id := make([]byte, 16)
				rand.Read(id)
				requestID = hex.EncodeToString(id)
				c.Set(_RequestIDKey, requestID)
				c.Writer.Header().Set("X-Request-Id", requestID)
			}
		}
		context, exists := c.Get(_ContextKey)
		var ctx *Context
		if exists {
			ctx = context.(*Context)
		} else {
			ctx = &Context{
				Context:   c,
				requestID: strings.ReplaceAll(requestID, "-", ""),
			}
			c.Set(_ContextKey, ctx)
		}

		defer func() {
			if err := recover(); err != nil {
				if log != nil {
					stack := stack(1)
					ctx.Errorf("[Recovery] panic recovered:\n%s\n%s", err, stack)
				}
				c.JSON(500, err)
			}
		}()
		if c.GetBool("X-Dumping") {
			h(ctx)
			return
		}
		c.Set("X-Dumping", true)

		dump, err := httputil.DumpRequest(c.Request, false)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Infof("Begin API Handler: %s------------------", dump)

		if c.Request.Body != http.NoBody && c.Request.Body != nil {
			var buf bytes.Buffer
			if _, err = buf.ReadFrom(c.Request.Body); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if err = c.Request.Body.Close(); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			ctx.Infof("Request: %s\n\n------------------", buf.String())
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
		}

		responseDumper := &responseDumper{ResponseWriter: c.Writer, outBuffer: bytes.NewBuffer(nil)}
		c.Writer = responseDumper

		h(ctx)
		ctx.Infof("End API Handler: \n%s\n------------------", responseDumper.bytes())
	}
}

// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
// For example, all the routes that use a common middlware for authorization could be grouped.
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		routerGroup: group.routerGroup.Group(relativePath, wrapHandlers(handlers)...),
	}
}

// BasePath get base path
func (group *RouterGroup) BasePath() string {
	return group.routerGroup.BasePath()
}

// Use adds middleware to the group, see example code in github.
func (group *RouterGroup) Use(handlers ...HandlerFunc) *RouterGroup {
	group.routerGroup.Use(wrapMiddlewares(handlers)...)
	return group
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in github.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) *RouterGroup {
	mode := gin.Mode()
	gin.SetMode(gin.ReleaseMode)
	group.routerGroup.Handle(httpMethod, relativePath, wrapHandlers(handlers)...)
	gin.SetMode(mode)

	debugPrintRoute(httpMethod, joinPaths(group.BasePath(), relativePath), handlers)
	return group
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("POST", relativePath, handlers...)
	return group
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("GET", relativePath, handlers...)
	return group
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("DELETE", relativePath, handlers...)
	return group
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("PATCH", relativePath, handlers...)
	return group
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("PUT", relativePath, handlers...)
	return group
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("OPTIONS", relativePath, handlers...)
	return group
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("HEAD", relativePath, handlers...)
	return group
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	group.Handle("GET", relativePath, handlers...)
	group.Handle("POST", relativePath, handlers...)
	group.Handle("PUT", relativePath, handlers...)
	group.Handle("PATCH", relativePath, handlers...)
	group.Handle("HEAD", relativePath, handlers...)
	group.Handle("OPTIONS", relativePath, handlers...)
	group.Handle("DELETE", relativePath, handlers...)
	group.Handle("CONNECT", relativePath, handlers...)
	group.Handle("TRACE", relativePath, handlers...)
	return group
}

// StdHandle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in github.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) StdHandle(httpMethod, relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	mode := gin.Mode()
	gin.SetMode(gin.ReleaseMode)
	group.routerGroup.Handle(httpMethod, relativePath, wrapStdHandlers(handlers)...)
	gin.SetMode(mode)

	debugPrintStdRoute(httpMethod, joinPaths(group.BasePath(), relativePath), handlers)
	return group
}

// StdPOST is a shortcut for router.Handle("POST", path, handle).
func (group *RouterGroup) StdPOST(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("POST", relativePath, handlers...)
	return group
}

// StdGET is a shortcut for router.Handle("GET", path, handle).
func (group *RouterGroup) StdGET(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("GET", relativePath, handlers...)
	return group
}

// StdDELETE is a shortcut for router.Handle("DELETE", path, handle).
func (group *RouterGroup) StdDELETE(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("DELETE", relativePath, handlers...)
	return group
}

// StdPATCH is a shortcut for router.Handle("PATCH", path, handle).
func (group *RouterGroup) StdPATCH(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("PATCH", relativePath, handlers...)
	return group
}

// StdPUT is a shortcut for router.Handle("PUT", path, handle).
func (group *RouterGroup) StdPUT(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("PUT", relativePath, handlers...)
	return group
}

// StdOPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (group *RouterGroup) StdOPTIONS(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("OPTIONS", relativePath, handlers...)
	return group
}

// StdHEAD is a shortcut for router.Handle("HEAD", path, handle).
func (group *RouterGroup) StdHEAD(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("HEAD", relativePath, handlers...)
	return group
}

// StdAny registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (group *RouterGroup) StdAny(relativePath string, handlers ...StdHandlerFunc) *RouterGroup {
	group.StdHandle("GET", relativePath, handlers...)
	group.StdHandle("POST", relativePath, handlers...)
	group.StdHandle("PUT", relativePath, handlers...)
	group.StdHandle("PATCH", relativePath, handlers...)
	group.StdHandle("HEAD", relativePath, handlers...)
	group.StdHandle("OPTIONS", relativePath, handlers...)
	group.StdHandle("DELETE", relativePath, handlers...)
	group.StdHandle("CONNECT", relativePath, handlers...)
	group.StdHandle("TRACE", relativePath, handlers...)
	return group
}

func wrapStdHandlers(handlers []StdHandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		ginHandlers[i] = wrapStdHandler(h)
	}
	return ginHandlers
}

func wrapMiddlewares(handlers []HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		ginHandlers[i] = wrapMiddleware(h)
	}
	return ginHandlers
}

func wrapMiddleware(h HandlerFunc) gin.HandlerFunc {

	return func(c *gin.Context) {

		requestID := c.GetString(_RequestIDKey)
		if len(requestID) == 0 {
			if requestID = c.Request.Header.Get("X-Request-Id"); len(requestID) > 0 {
				c.Set(_RequestIDKey, requestID)
				c.Writer.Header().Set("X-Request-Id", requestID)
			} else {
				id := make([]byte, 16)
				rand.Read(id)
				requestID = hex.EncodeToString(id)
				c.Set(_RequestIDKey, requestID)
				c.Writer.Header().Set("X-Request-Id", requestID)
			}
		}

		context, exists := c.Get(_ContextKey)
		var ctx *Context
		if exists {
			ctx = context.(*Context)
		} else {
			ctx = &Context{
				Context:   c,
				requestID: strings.ReplaceAll(requestID, "-", ""),
			}
			c.Set(_ContextKey, ctx)
		}
		h(ctx)
	}
}

func wrapStdHandler(h StdHandlerFunc) gin.HandlerFunc {
	return wrapHandler(func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := stack(4)
				c.Errorf("panic:\n%s\n%s", err, stack)

				c.Set("code", 500)
				c.JSON(200, gin.H{
					"code":    500,
					"message": fmt.Sprint(err),
					"data":    nil,
				})
			}
		}()

		h(c)
		//code, message, data := h(c)
		//if code == 0 && message == "" {
		//	message = "success"
		//}
		//if code != 0 && code != 999999 {
		//	c.Error("error message: " + message)
		//	c.Error("error message: " + message)
		//}
		//c.Set("code", code)
		//c.Set("message", message)
		////traverseTimeField(c, data)
		//c.JSON(200, gin.H{
		//	"code":    code,
		//	"message": message,
		//	"data":    data,
		//})
	})
}

// Last returns the last handler in the chain. ie. the last handler is the main own.
func last(c []HandlerFunc) HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

func debugPrintRoute(httpMethod, absolutePath string, handlers []HandlerFunc) {
	nuHandlers := len(handlers)
	handlerName := nameOfFunction(last(handlers))
	fmt.Printf("[GIN] %-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

func debugPrintStdRoute(httpMethod, absolutePath string, handlers []StdHandlerFunc) {
	nuHandlers := len(handlers)
	handlerName := nameOfFunction(lastStd(handlers))
	fmt.Printf("[GIN] %-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
}

// Last returns the last handler in the chain. ie. the last handler is the main own.
func lastStd(c []StdHandlerFunc) StdHandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}
