package main

import (
	"fmt"
	"github.com/Chairou/toolbox/conf"
	g "github.com/Chairou/toolbox/gin"
	"github.com/Chairou/toolbox/timeformat"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	gorm "gorm.io/gorm"
	"io"
	"net/http"
	"os"
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
}

func (Catalog) TableName() string { return "t_catalog" }

func main() {

	g.SetRouterRegister(func(group *g.RouterGroup) {
		routerGroup := group.Group("/api")
		routerGroup.StdGET("get", get)
		routerGroup.StdPOST("postBody", postBody)
		routerGroup.StdGET("catalog", getDataFromMysql)
		routerGroup.StdGET("ping", func(c *g.Context) {
			g.WriteRetJson(c, 0, nil, "pong")
		})
	})
	r := g.NewServer("dev", "srv.log")

	fmt.Println("start server at *:80")
	err := r.Run(":80")
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
	c.String(http.StatusOK, string(body))
}

func getDataFromMysql(c *g.Context) {
	mysqlInit()
	parConstruct := map[string]*g.ParamConstruct{
		"id":          {FieldName: "id", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"englishName": {FieldName: "englishName", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"chineseName": {FieldName: "chineseName", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "like"},
		"status":      {FieldName: "status", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: "="},
		"createtime":  {FieldName: "createtime", DefaultValue: "", CheckValue: nil, Need: false, Link: "and", Symbol: ">="},
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

	var catalogs []Catalog
	DbConn.Find(&catalogs).Where(strCondition, args).Order(orderStr).Offset(offset).Limit(limit)
	c.RetJson(0, catalogs, "ok")

}

func mysqlInit() {
	// 初始化mysql
	conf.LoadAllConf(&config)
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
		panic("failed to connect database," + dsn)
	}
}
