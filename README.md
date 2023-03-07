# toolbox
[English version](https://github.com/Chairou/toolbox/wiki#toolbox) 

A simple and minimally dependent Golang toolbox and development library. Contact email <chair.ou#gmail.com>.

一个简单, 尽量少依赖的golang工具箱和开发库,联系email <chair.ou#gmail.com>
 

# 创作动机
近二十年的技术工作, 发现自己并没有太多的积累,所以, 我开始了这个项目,
认认真真的做一个工具库出来, 一来技术有积累有进步, 二来方便大家一起使用, 三来一旦被离职也有底气.
请大家重视开源工作的原因也在此三点.
江湖路远, 各位珍重

## 目标:

1. 简单, 开箱即用
2. 尽量函数化, 减少状态
3. 测试用例完备

## 目录:
### httphelper
    用途: 封装好HTTP的GET和POST,封装常用转义, 以及处理返回的JSON串
    测试用例: go test -v http_test.go http.go helper.go result.go (注意,暂未做mock,自己改为能用的IP)
    主要函数:
        func GET(url string) Helper //发送GET请求
        func PostJSON(url string, body interface{}) Helper //发送POST请求, 内容为JSON格式的body
        func UrlPathEscape(url string) string //对URL进行转义
        func (p *baseResult) BaseResult() *baseResult // BaseResult 返回Http请求的基本结果，包含Status和Body
        func (p *jsonResult) Bind(object interface{}, path ...interface{}) error 
        // Bind 将返回值存储到Object中
        注意, 测试用例中有对返回字串和JSON串的处理方法

### logger
    用途: 最简单的打日志, 支持日志分割,日志等级调整
    测试用例: go test -v log_test.go logger.go
    主要函数:
        func NewLogPool(fileName string) (*logPool, error) // 生成日志实例
        func GetLogPool(fileName string) (*logPool, error) // 获得日志实例
        func (c *logPool) Debugf(format string, v ...any) // 写debug日志
        func (c *logPool) Debugln(v ...any) // 写debug日志
        func (c *logPool) Infof(format string, v ...any) // 写INFO日志
        func (c *logPool) Infoln(v ...any) // 写INFO日志
        func (c *logPool) Errorf(format string, v ...any) //写错误日志
        func (c *logPool) Errorln(v ...any) // 写错误日志
        func (c *logPool) SetLevel(level int) error // 设置日志等级

### util/redis
    用途: 操作多个redis池
    测试用例: go test -v redis_test.go redis.go (注意, 需要自行搭建redis服务, 后继做mock)
    主要函数:
        func NewRedis(name string, addr string, passwd string) *RedisPool // 生成新的redis实例并放入Pool中
        func GetRedisPool(name string) (*RedisPool, error) // 获取redis实例
        func GetRedisByName(name string) *RedisPool // 获取redis实例, 忽略错误
        func (c *RedisPool) Get(key string) (string, error) // redis GET
        func (c *RedisPool) SilenceGet(key string) string // 静默获取kv, 不会返回错误, 有问题只返回空字串
        func (c *RedisPool) HGet(key string, subKey string) (string, error) // redis HGET
        func (c *RedisPool) Set(key string, val string) (string, error) // redis SET
        func (c *RedisPool) HSet(key string, subKey string, val string) (int64, error) // redis HSET
        func (c *RedisPool) Del(key string) (int64, error) // redis DEL
        func (c *RedisPool) Do(commandName string, args ...interface{}) (interface{}, error) // redis DO 通用接口
        func (c *RedisPool) ClosePool() error // 关闭连接池, 释放sync.Map
        func (c *RedisPool) Expired(key string, seconds int) (int64, error) // redis expire
        func (c *RedisPool) Ttl(key string) (int64, error) // redis TTL
        func (c *RedisPool) HMGet(key string, values ...string) ([]string, error) // redis HMGET
        func (c *RedisPool) HMSet(key string, kv map[string]string) (string, error) // redis HMSET
        func (c *RedisPool) HGetAll(key string) (map[string]string, error) // redis HGETALL
        func (c *RedisPool) HSetEX(key, field string, value interface{}, expire int) (int64, error) 
        // redis HSET and expire
        func (c *RedisPool) Increment(key string) (int64, error) // redis INCR
        func (c *RedisPool) LIndex(key string, index int) (string, error) // redis LINDEX
        func (c *RedisPool) LLen(key string) (int64, error) //redis LLEN
        func (c *RedisPool) LPop(key string) (string, error) // redis LPOP
        func (c *RedisPool) LPush(key string, values ...interface{}) (int64, error) // redis LPUSH
        func (c *RedisPool) LPushX(key string, values ...interface{}) (int64, error) // redis LPUSHX
        func (c *RedisPool) LRem(key string, count int, value string) (int64, error) // redis LREM
        func (c *RedisPool) LSet(key, value string, index int) (int64, error) // redis LSET
        func (c *RedisPool) LTrim(key string, start, stop int) (string, error) // redis LTRIM
        func (c *RedisPool) MGet(keys ...string) ([]string, error) // REDIS MGET
        func (c *RedisPool) MSet(pairs ...interface{}) (string, error) / REDIS MSET
        func (c *RedisPool) Ping() (string, error) // REDIS PING
        func (c *RedisPool) SetEX(key string, value interface{}, expire int) (string, error) // REDIS SETEX
        func (c *RedisPool) SetNX(key string, value interface{}, expire int) (int64, error) // REDIS SETNX
        func (c *RedisPool) LRange(key string, start, stop int64) ([]string, error) // REDIS LRANGE

### util/cal
    用途: 计算日期和IP范围
    测试用例: go test -v calDate_test.go calDate.go calIP_test.go calIP.go
    主要函数:
        // 获取两个时间相差的天数,小时数,分钟数,秒数, 用法见函数注释和测试用例
        func GetDiffTime(previousTime, laterTime interface{}, flag int) (int64, error) {
        func GetFirstAndLastDateOfWeek(date time.Time) // 获取当天所在周的周一和周日时间
        func SubNetMaskToLen(netmask string) (int, error) // 获取子网掩码位数
        func LenToSubNetMask(subnet int) string // 从掩码位数转换为点十分制的子网掩码
        func GetCidrIpRange(cidr string) (first string, broadcast string) // 获得子网第一个IP地址和广播地址

### util/check
    用途: 数据合法性检查
    测试用例:  go test -v valiad_test.go checkValiad.go
    主要函数:
        func FilteredSQLInject(toMatchStr string) bool // 判断是否有SQL注入
        func IsNumeric(val interface{}) bool // 判断是否为数字
        func CheckEmail(email string) (err error) // 检查是否合法email地址
        func CheckMobile(mobile string) bool // 检查是否手机号码
        func IsValidIDCardCheckSum(idCard string) bool // 检查是否合法身份证号
        func CheckField(field string) (err error) // 检查合法输入, 白名单, 汉字, 数字, 字母,下划线,点
        func CheckIP(ip string) bool // 检查是否合法的IPV4和IPV6地址

### util/conv
    用途: 类型转换
    测使用例: go test -v conv_test.go conv.go
    主要函数:
        func GbkToUtf8(s []byte) ([]byte, error) // GBK转UTF-8
        func Utf8ToGbk(s []byte) ([]byte, error) // UTF-8转GBK
        func String(val interface{}) string // 所有类型转为string
        func Int64(val interface{}) (int64, bool) // 转为int64
        func Uint64(val interface{}) (uint64, bool) // 转为uint64
        func Int(val interface{}) (int, bool) // 转为int
        func Uint(val interface{}) (uint, bool) // 转为uint
        func Float64(val interface{}) (float64, bool) // 转为float64
        func Bool(val interface{}) (bool, bool) // 转为bool
        func IsNil(val interface{}) bool // 判断是否为nil
        func Time(val interface{}) (time.Time, bool) // 把20060102, 2006-01-02, 2006-01-02 15:04:05
        这三种类型的string转为time类型
        func TimePtr(val interface{}) *time.Time // 把上述3种类型转为time指针
        func StringToArray(ext string) (array []string, err error) //"-a 123 -b hello" ---> ["-a","123","-b","hello"]
        func StringToMap(ext string) (map[string]string, error) // "-a 123 -b hello" ---> {"-a":"123","-b":"hello"}
        func JsonToArray(ext string) ([]string, error) 
        // "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> ["-a","123","-b","hello"]
        func JsonToString(ext string) (string, error) // "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> "-a 123 -b hello"
        func StructToMap(in interface{}, tagName string) (map[string]interface{}, error) // 结构体转为Map

### util/crypt
    用途: 加密
    测试用例: go test -v  crypt_test.go crypt.go
    主要函数: 
        func AesEncrypt2(orig string, key string) (string, error) // 加密
        func AesDecrypt2(cryted string, key string) (string, error) // 解密

## util/encode
    用途: 编码
    测试用例:  go test -v encode_test.go encode.go
    主要函数:
        func FlateCompress(origData []byte) (result []byte, err error) // 压缩
        func FlateUnCompress(compressData []byte) (result []byte, err error) // 解压
        func Base64Encode(origData []byte) (result string) // 转base64
        func Base64Decode(encodedData string) (result []byte, err error) // 从base64转回
        func MD5(origData []byte) (result string) // 转MD5
        func MD5File(fileName string) (result string, err error) // 对文件内容转MD5
        func Sha1(origData []byte) (result string) // 转SHA1
        func Sha1File(fileName string) (result string, err error) // 对文件内容转SHA1
        func Sha256(origData []byte) (result string) // 转SHA256
        func Sha256File(fileName string) (result string, err error) // 对文件内容转SHA256
        func Sha512(origData []byte) (result string) // 转SHA512
        func Sha512File(fileName string) (result string, err error) // 对文件内容转SHA512
        func PKCS5Padding(plaintext []byte, blockSize int) []byte // 按PKCS5格式填充明文

### util/listopt
    用途: list的各类操作
    测试用例: go test -v list_test.go list_opt.go
    主要函数:
        func SplitList(arr []string, num int64) [][]string // 平均分割一个list到num个list里
        func RemoveRepeatedElement(slice interface{}) []interface{} // 移除数组中重复的元素
        func RemoveDuplicateString(languages []string) []string // 移除数组中重复的string
        func RemoveDuplicateInt(languages []int) []int // 移除数组中重复的int
        func DeleteString(strList []string, delStr string) []string // 删除list中指定的string
        func IntersectStr(slice1, slice2 []string) []string // 求交集
        func UnionStr(slice1, slice2 []string) []string // 求并集
        func DifferenceStr1(slice1, slice2 []string) []string // 求差集
        func DifferenceStr2(slice1 []string, slice2 []string) []string // 求差集
        func In(strList []string, target string) bool // 判断string是否在list内
        func ReverseStr(arr []string) []string // 反序输出list

### util/mapsort
    用途: 对struct和map进行排序
    测试用例: go test -v mapsort_test.go
    主要函数:
        func TestSortSample(t *testing.T) // 对struct排序的例子
        func TestRankByWordCount(t *testing.T) // 对map排序的例子

### util/bitmap
    用途: 实现bitmap做状态记录
    测试用例: go test -v bitmap_test.go
    主要函数:
        func NewBitMap(name string, n uint64) *BitMapStruct // 生成bitmap对象
        func GetBitMap(name string) *BitMapStruct // 获取bitmap实例
        func (bt *BitMapStruct) Set(n uint64) error // 设置对应位置的bit为1
        func (bt *BitMapStruct) MSet(n ...uint64) // 批量设置bits
        func (bt *BitMapStruct) Del(n uint64) error // 设置对应bit为0
        func (bt *BitMapStruct) MDel(elements ...uint64) //批量设置对应bit为0
        func (bt *BitMapStruct) IsExist(n uint64) bool // 返回是否置位
        func (bt *BitMapStruct) MExist(elements ...uint64) map[uint64]bool // 批量判断是否置位
        func (bt *BitMapStruct) PrintAllBits() // 打印整个bitmap
        func (bt *BitMapStruct) Clean() // 清零整个bitmap
        func (bt *BitMapStruct) Destroy() // Destroy 删除当前bitmap



## 下一步计划:
    1. 完成英文文档(完成)
    2. 所有对外函数都加上swagger格式的注释(进行中)
    3. 把mock做上