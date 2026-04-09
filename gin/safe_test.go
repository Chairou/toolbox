package gin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter 创建一个用于测试的路由，注册 SafeCheck 中间件
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/test", func(c *gin.Context) {
		// 包装为自定义 Context
		ctx := &Context{Context: c}
		SafeCheck(ctx)
		// 如果 SafeCheck 没有 Abort，则返回正常响应
		if !c.IsAborted() {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		}
	})

	return r
}

// TestSafeCheck_PostWithSelectInjection 测试 POST 请求 body 中包含 select ... from 的 SQL 注入
func TestSafeCheck_PostWithSelectInjection(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"name":  "normal_user",
		"query": "select * from users where id = 1",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusForbidden, w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应 body 失败: %v", err)
	}
	if resp["message"] != "访问被禁止" {
		t.Errorf("期望响应消息为 '访问被禁止'，实际得到 '%s'", resp["message"])
	}
	t.Logf("SQL注入检测成功，响应: %s", w.Body.String())
}

// TestSafeCheck_PostWithUnionSelectInjection 测试 POST 请求 body 中包含 union select 的 SQL 注入
func TestSafeCheck_PostWithUnionSelectInjection(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"id":    1,
		"input": "1 union select username, password from admin--",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusForbidden, w.Code)
	}
	t.Logf("UNION SELECT 注入检测成功，响应: %s", w.Body.String())
}

// TestSafeCheck_PostWithNormalData 测试 POST 请求 body 中包含正常数据，不应被拦截
func TestSafeCheck_PostWithNormalData(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   25,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际得到 %d，正常数据不应被拦截", http.StatusOK, w.Code)
	}
	t.Logf("正常数据通过检测，响应: %s", w.Body.String())
}

// TestSafeCheck_PostWithDeleteInjection 测试 POST 请求 body 中包含 delete from 的 SQL 注入
func TestSafeCheck_PostWithDeleteInjection(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"action": "delete from users where 1=1",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusForbidden, w.Code)
	}
	t.Logf("DELETE 注入检测成功，响应: %s", w.Body.String())
}

// TestSafeCheck_PostWithDropTableInjection 测试 POST 请求 body 中包含 drop table 的 SQL 注入
func TestSafeCheck_PostWithDropTableInjection(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"sql": "drop table users",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusForbidden, w.Code)
	}
	t.Logf("DROP TABLE 注入检测成功，响应: %s", w.Body.String())
}

// TestSafeCheck_PostWithCommentInjection 测试 POST 请求 body 中包含 SQL 注释的注入
func TestSafeCheck_PostWithCommentInjection(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"input": "admin'--",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusForbidden, w.Code)
	}
	t.Logf("SQL注释注入检测成功，响应: %s", w.Body.String())
}
