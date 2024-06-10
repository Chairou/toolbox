package main

import (
	"fmt"
	g "github.com/Chairou/toolbox/gin"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func main() {

	r := gin.Default()
	r.Use(g.SafeCheck())
	r.Use(g.ResponseRecorder())
	api := r.Group("/api")
	api.GET("/ping", func(c *gin.Context) {
		g.WriteRetJson(c, 0, nil, "pong")
	})
	r.GET("/get", get)
	r.POST("/postBody", postBody)
	r.POST("/upload", g.RecUploadFile)
	fmt.Println("starting http server")
	err := r.Run(":80")
	if err != nil {
		fmt.Println("RUN err:", err)
		return
	}
}

func get(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	params := make(map[string]string)
	for key, values := range queryParams {
		params[key] = values[0]
	}
	// 返回所有GET参数的JSON响应
	c.JSON(http.StatusOK, params)
}

func postBody(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, string(body))
}
