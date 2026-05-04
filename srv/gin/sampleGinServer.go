package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Chairou/toolbox/conf"
	g "github.com/Chairou/toolbox/gin"
	"github.com/Chairou/toolbox/timeformat"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	gorm "gorm.io/gorm"
)

type Config struct {
	Env          string `yaml:"env" json:"env" env:"env"`
	Version      int    `yaml:"version" json:"version" env:"version"`
	RedisName    string `yaml:"redis_name" json:"redis_name" env:"redis_name"`
	RedisHost    string `yaml:"redis_host" json:"redis_host" env:"redis_host"`
	RedisAuth    string `yaml:"redis_auth" json:"redis_auth" env:"redis_auth"`
	MysqlName    string `yaml:"mysql_name" json:"mysql_name" env:"mysql_name"`
	MysqlHost    string `yaml:"mysql_host" json:"mysql_host" env:"mysql_host"`
	MysqlUser    string `yaml:"mysql_user" json:"mysql_user" env:"mysql_user"`
	MysqlPass    string `yaml:"mysql_pass" json:"mysql_pass" env:"mysql_pass"`
	MysqlDb      string `yaml:"mysql_db" json:"mysql_db" env:"mysql_db"`
	MysqlCharSet string `yaml:"mysql_charset" json:"mysql_charset" env:"mysql_charset"`
	LogFileName  string `yaml:"log_file_name" json:"log_file_name" env:"log_file_name"`
	FileName     string `yaml:"fileName" json:"fileName" env:"fileName"`             // 日志路径和文件名 如./log/test.log
	Level        int    `yaml:"level" json:"level" env:"level"`                      // 日志级别，可选 DEBUG_LEVEL、INFO_LEVEL、ERROR_LEVEL
	MaxSizeMB    int    `yaml:"maxSizeMB" json:"maxSizeMB" env:"maxSizeMB"`          // 单个日志文件最大大小（MB）
	MaxBackups   int    `yaml:"maxBackups" json:"maxBackups" env:"maxBackups"`       // 最大保留的旧日志文件数量
	MaxAgeDay    int    `yaml:"maxAgeDay" json:"maxAgeDay" env:"maxAgeDay"`          // 旧日志文件最大保留天数
	Compress     int    `yaml:"compress" json:"compress" env:"compress"`             // 是否压缩旧日志文件，默认不压缩
	PrintConsole int    `yaml:"printConsole" json:"printConsole" env:"printConsole"` // 是否同时输出到控制台
}

type Catalog struct {
	ID          uint64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	EnglishName string          `gorm:"column:englishName;type:varchar(255);not null;comment:英文名" json:"englishName"`
	ChineseName string          `gorm:"column:chineseName;type:varchar(255);not null;collation:utf8mb4_0900_ai_ci;comment:中文名" json:"chineseName"`
	Status      int             `gorm:"column:status;type:int;not null;default:1;comment:1启用，2停用" json:"status"`
	CreateTime  timeformat.Time `gorm:"column:createTime;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createTime"`
	UpdateTime  timeformat.Time `gorm:"column:updateTime;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updateTime"`
	Updater     string          `gorm:"column:updater;type:varchar(100);not null;comment:更新人" json:"updater"`
}

var config = Config{}
var DbConn *gorm.DB

func init() {
	_ = os.Setenv("mysql_host", "127.0.0.1:3306")
	_ = os.Setenv("mysql_user", "root")
	_ = os.Setenv("mysql_pass", "root123456")
	_ = os.Setenv("fileName", "test.log")
	_ = os.Setenv("level", "0")
	_ = os.Setenv("maxSizeMB", "100")
	_ = os.Setenv("maxBackups", "10")
	_ = os.Setenv("MaxAgeDay", "31")
	_ = os.Setenv("compress", "0")
	_ = os.Setenv("printConsole", "0")
}

func (Catalog) TableName() string { return "t_catalog" }

func main() {
	conf.LoadAllConf(&config)
	mysqlInit()
	g.SetRouterRegister(func(group *g.RouterGroup) {
		routerGroup := group.Group("/api")
		routerGroup.StdGET("get", get)
		routerGroup.StdPOST("postBody", postBody)
		routerGroup.StdGET("catalog", getDataFromMysql)
		routerGroup.StdGET("ping", func(c *g.Context) {
			g.WriteRetJson(c, 0, nil, "pong")
		})
	})
	//r := g.NewServer("dev", "srv.log", nil)
	r := g.NewServerWithConf("dev", config, nil)

	fmt.Println("start server at *:8080")
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("RUN err:", err)
		return
	}
}

func get(c *g.Context) {
	queryParams := c.Request.URL.Query()
	params := make(map[string]string)
	for key, values := range queryParams {
		params[key] = values[0]
	}
	// 返回所有GET参数的JSON响应
	c.JSON(http.StatusOK, params)
}

func postBody(c *g.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.Request.Body.Close()
	c.String(http.StatusOK, string(body))
}

func getDataFromMysql(c *g.Context) {
	parConstruct := map[string]*g.ParamConstruct{
		"id":          {FieldName: "id", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"englishName": {FieldName: "englishName", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"chineseName": {FieldName: "chineseName", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "like"},
		"status":      {FieldName: "status", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"createtime":  {FieldName: "createTime", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: ">="},
		"endtime":     {FieldName: "createTime", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "<="},
		"searchKey":   {FieldName: "englishName|chineseName", DefaultValue: "", CheckValue: nil, Need: false, Link: "and"},
		"orderBy":     {FieldName: "", DefaultValue: "id|desc", CheckValue: nil, Need: false},
	}

	strCondition, args, orderStr, err := c.GetConditionByParam(parConstruct)
	if err != nil {
		c.RetJson(-101, nil, "param err: ", err)
		return
	}
	//获取分页参数:PageIndex,PageSize, 不传的话默认10000条数据
	offset, limit, err := c.GetLimit()
	if err != nil {
		c.RetJson(-101, nil, "param err: ", err)
		return
	}

	var catalogs []Catalog
	DbConn.Where(strCondition, args...).Order(orderStr).Offset(offset).Limit(limit).Find(&catalogs)
	c.RetJson(0, catalogs, "ok")

}

func mysqlInit() {
	// 初始化mysql
	host := config.MysqlHost
	user := config.MysqlUser
	password := config.MysqlPass
	db := config.MysqlDb
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, db)

	fmt.Printf("config: %#v\n", config)
	var err error
	DbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
	})
	if err != nil {
		log.Fatal("failed to connect database," + host)
	}
}
