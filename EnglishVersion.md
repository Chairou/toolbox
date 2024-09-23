# toolbox
A simple and minimally dependent Golang toolbox and development library. Contact email <chair.ou#gmail.com>.

[中文版本](https://github.com/Chairou/toolbox#readme)

# Motivation
After nearly 20 years of technical work, I found that I didn't have much accumulation. Therefore, I started this project to create a tool library seriously. First, I can accumulate and improve my skills. Second, it's convenient for everyone to use together. Third, once I leave the job, I have confidence. Please also focus on the three reasons for open source work. The road ahead is long, and take care.

## Goals:
    Simple, easy to use out of the box
    Try to functionize and reduce state
    Complete test cases

## Directory:

### httphelper
    Purpose: Encapsulate HTTP's GET and POST, encapsulate common escape and process returned JSON strings.
    Test Case: go test -v http_test.go http.go helper.go result.go 
    (note that mock has not been done yet, change it to an available IP yourself)
    Main Function:
    func GET(url string) Helper //send GET request
    func PostJSON(url string, body interface{}) Helper //send POST request, with JSON format body
    func UrlPathEscape(url string) string //escape URL
    // BaseResult returns the basic result of the Http request, including Status and Body
    func (p *baseResult) BaseResult() *baseResult 
    func (p *jsonResult) Bind(object interface{}, path ...interface{}) error
    //Bind stores the return value in Object
    Note that there are methods for processing return strings and JSON strings in the test case.

### logger
    Purpose: The simplest way to log, supporting log segmentation and log level adjustment.
    Test Case: go test -v log_test.go logger.go
    Main Function:
    func NewLogPool(fileName string) (*logPool, error) //generate log instance
    func GetLogPool(fileName string) (*logPool, error) //get log instance
    func (c *logPool) Debugf(format string, v ...any) //write debug logs
    func (c *logPool) Debugln(v ...any) //write debug logs
    func (c *logPool) Infof(format string, v ...any) //write INFO logs
    func (c *logPool) Infoln(v ...any) //write INFO logs
    func (c *logPool) Errorf(format string, v ...any) //write error logs
    func (c *logPool) Error(v ...any) //write error logs
    func (c *logPool) SetLevel(level int) error //set log level

### util/redis
    Purpose: operate multiple redis pools
    Test case: go test -v redis_test.go redis.go 
    (Note,you need to build a redis service yourself, and then make a mock)
    main function:
        //Generate a new redis instance and put it in the Pool
        func NewRedis(name string, addr string, passwd string) *RedisPool 
        //get redis instance
        func GetRedisPool(name string) (*RedisPool, error)
        // get redis instance, ignore errors
        func GetRedisByName(name string) *RedisPool
        // Get kv silently, no error will be returned but empty string, if there is a problem.
        func (c *RedisPool) SilenceGet(key string) string
        func (c *RedisPool) Get(key string) (string, error) // redis GET
        func (c *RedisPool) HGet(key string, subKey string) (string, error) // redis HGET
        func (c *RedisPool) Set(key string, val string) (string, error) // redis SET
        func (c *RedisPool) HSet(key string, subKey string, val string) (int64, error) // redis HSET
        func (c *RedisPool) Del(key string) (int64, error) // redis DEL
        // redis DO generic interface
        func (c *RedisPool) Do(commandName string, args ...interface{}) (interface{}, error) 
        func (c *RedisPool) ClosePool() error // close the connection pool, release sync.Map
        func (c *RedisPool) Expired(key string, seconds int) (int64, error) // redis expire
        func (c *RedisPool) Ttl(key string) (int64, error) // redis TTL
        func (c *RedisPool) HMGet(key string, values ...string) ([]string, error) // redis HMGET
        func (c *RedisPool) HMSet(key string, kv map[string]string) (string, error) // redis HMSET
        func (c *RedisPool) HGetAll(key string) (map[string]string, error) // redis HGETALL
        // redis HSET and expire
        func (c *RedisPool) HSetEX(key, field string, value interface{}, expire int) (int64, error)
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
     Purpose: Calculate date and IP range
     Test cases: go test -v calDate_test.go calDate.go calIP_test.go calIP.go
     main function:
        // Get the days, hours, minutes, and seconds of the difference between two times, 
        // see function comments and test cases for usage
        func GetDiffTime(previousTime, laterTime interface{}, flag int) (int64, error) 
        // Get the Monday and Sunday time of the current week
        func GetFirstAndLastDateOfWeek(date time.Time) 
        func SubNetMaskToLen(netmask string) (int, error) // Get the number of subnet mask bits
        // Convert from the number of mask digits to the subnet mask of dot-ten system
        func LenToSubNetMask(subnet int) string 
        // Get the first IP address and broadcast address of the subnet
        func GetCidrIpRange(cidr string) (first string, broadcast string) 


### util/check
    Purpose: data legality check
    Test case: go test -v valiad_test.go checkValiad.go
    main function:
        func FilteredSQLInject(toMatchStr string) bool // determine whether there is SQL injection
        func IsNumeric(val interface{}) bool // determine whether it is a number
        func CheckEmail(email string) (err error) // Check if the email address is valid
        func CheckMobile(mobile string) bool // Check if the mobile number
        func IsValidIDCardCheckSum(idCard string) bool // Check whether the ID card number is valid
        // Check legal input, whitelist, Chinese characters, numbers, letters, underscores, dots    
        func CheckField(field string) (err error) 
        func CheckIP(ip string) bool // Check whether the IPV4 and IPV6 addresses are legal

## util/conv
    Purpose: type conversion
    Test example: go test -v conv_test.go conv.go
    main function:
        func GbkToUtf8(s []byte) ([]byte, error) // GBK to UTF-8
        func Utf8ToGbk(s []byte) ([]byte, error) // UTF-8 to GBK
        func String(val interface{}) string // convert all types to string
        func Int64(val interface{}) (int64, bool) // convert to int64
        func Uint64(val interface{}) (uint64, bool) // convert to uint64
        func Int(val interface{}) (int, bool) // convert to int
        func Uint(val interface{}) (uint, bool) // convert to uint
        func Float64(val interface{}) (float64, bool) // convert to float64
        func Bool(val interface{}) (bool, bool) // convert to bool
        func IsNil(val interface{}) bool // determine whether it is nil
        // put 20060102, 2006-01-02, 2006-01-02 15:04:05 These three types of 
        // strings are converted to time types    
        func Time(val interface{}) (time.Time, bool)
        func TimePtr(val interface{}) *time.Time // convert the above 3 types into time pointers
        //"-a 123 -b hello" ---> ["-a","123","-b","hello"]
        func StringToArray(ext string) (array []string, err error) 
        // "-a 123 -b hello" ---> {"-a":"123","-b":"hello"}
        func StringToMap(ext string) (map[string]string, error) 
        // "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> ["-a", "123", "-b", "hello"]        
        func JsonToArray(ext string) ([]string, error)
        // "{\"-a\":\"123\",\"-b\":\"hello\"}" ---> "-a 123 -b hello"
        func JsonToString(ext string) (string, error) 
        // Convert structure to Map
        func StructToMap(in interface{}, tagName string) (map[string]interface{}, error) 


### util/crypt
    Purpose: encryption
    Test case: go test -v crypt_test.go crypt.go
    main function:
        func AesEncrypt2(orig string, key string) (string, error) // encryption
        func AesDecrypt2(cryted string, key string) (string, error) // Decrypt

### util/encode
     Usage: Coding
     Test case: go test -v encode_test.go encode.go
     main function:
         func FlateCompress(origData []byte) (result []byte, err error) // compression
         func FlateUnCompress(compressData []byte) (result []byte, err error) // decompress
         func Base64Encode(origData []byte) (result string) // convert to base64
         func Base64Decode(encodedData string) (result []byte, err error) // convert back from base64
         func MD5(origData []byte) (result string) // convert to MD5
         func MD5File(fileName string) (result string, err error) // Convert file content to MD5
         func Sha1(origData []byte) (result string) // convert to SHA1
         func Sha1File(fileName string) (result string, err error) // Convert file content to SHA1
         func Sha256(origData []byte) (result string) // convert to SHA256
         func Sha256File(fileName string) (result string, err error) // Convert file content to SHA256
         func Sha512(origData []byte) (result string) // convert to SHA512
         func Sha512File(fileName string) (result string, err error) // Convert file content to SHA512
         func PKCS5Padding(plaintext []byte, blockSize int) []byte // pad plaintext in PKCS5 format


### util/listopt
    Purpose: various operations of list
    Test case: go test -v list_test.go list_opt.go
    main function:
        func SplitList(arr []string, num int64) [][]string // averagely split a list into num lists
        func RemoveRepeatedElement(slice interface{}) []interface{} // Remove repeated elements in the array
        func RemoveDuplicateString(languages []string) []string // Remove duplicate strings from the array
        func RemoveDuplicateInt(languages []int) []int // remove the duplicate int in the array
        func DeleteString(strList []string, delStr string) []string // Delete the string specified in the list
        func IntersectStr(slice1, slice2 []string) []string // find intersection
        func UnionStr(slice1, slice2 []string) []string // find the union
        func DifferenceStr1(slice1, slice2 []string) []string // difference set
        func DifferenceStr2(slice1 []string, slice2 []string) []string // find difference set
        func In(strList []string, target string) bool // determine whether the string is in the list
        func ReverseStr(arr []string) []string // Output list in reverse order

### util/mapsort
    Purpose: sort struct and map
    Test case: go test -v mapsort_test.go
    main function:
        func TestSortSample(t *testing.T) // example of sorting struct
        func TestRankByWordCount(t *testing.T) // example of sorting map

### util/bitmap
    Purpose: Realize bitmap for status record
    Test case: go test -v bitmap_test.go
    main function:
        func NewBitMap(name string, n uint64) *BitMapStruct // generate bitmap object
        func GetBitMap(name string) *BitMapStruct // Get bitmap instance
        func (bt *BitMapStruct) Set(n uint64) error // set the corresponding bit to 1
        func (bt *BitMapStruct) MSet(n ...uint64) // set bits in batches
        func (bt *BitMapStruct) Del(n uint64) error // set the corresponding bit to 0
        func (bt *BitMapStruct) MDel(elements ...uint64) //Batch setting corresponding bit to 0
        func (bt *BitMapStruct) IsExist(n uint64) bool // return whether set
        func (bt *BitMapStruct) MExist(elements ...uint64) map[uint64]bool // Batch judgment whether set
        func (bt *BitMapStruct) PrintAllBits() // print the whole bitmap
        func (bt *BitMapStruct) Clean() // clear the entire bitmap
        func (bt *BitMapStruct) Destroy() // Destroy deletes the current bitmap


## Next steps:
     1. Complete the English document (done)
     2. All external functions are commented in swagger format
     3. Make the mock
