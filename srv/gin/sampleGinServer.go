package main

import (
	"fmt"
	g "github.com/Chairou/toolbox/gin"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func main() {

	g.SetRouterRegister(func(group *g.RouterGroup) {
		routerGroup := group.Group("/api")
		routerGroup.StdGET("get", get)
		routerGroup.StdPOST("postBody", postBody)
		routerGroup.StdGET("ping", func(c *g.Context) {
			g.WriteRetJson(c, 0, nil, "pong")
		})
	})
	r := g.NewServer()

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
