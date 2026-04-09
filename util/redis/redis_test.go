package redis

import (
	"sync"
	"testing"
	"time"
)

// go test -v redis_test.go redis.go

const (
	testAddr   = "127.0.0.1:6379"
	testPasswd = "chairou"
	testPool   = "redis_test"
)

// getTestRedis 获取测试用的 Redis 实例
func getTestRedis(t *testing.T) *RdPool {
	t.Helper()
	inst := NewRedis(testPool, testAddr, testPasswd)
	if inst == nil {
		t.Fatal("NewRedis 返回 nil")
	}
	return inst
}

// ========== 1. NewRedis 竞态条件修复 + Name 字段初始化 ==========

// TestNewRedis_NameFieldInitialized 验证 Name 字段被正确初始化
func TestNewRedis_NameFieldInitialized(t *testing.T) {
	name := "test_name_init"
	inst := NewRedis(name, testAddr, testPasswd)
	defer func() {
		_ = inst.ClosePool()
	}()

	if inst.Name != name {
		t.Errorf("期望 Name=%q, 实际 Name=%q", name, inst.Name)
	}
}

// TestNewRedis_LoadOrStore 验证并发调用 NewRedis 返回同一实例（竞态修复）
func TestNewRedis_LoadOrStore(t *testing.T) {
	name := "test_load_or_store"
	// 先清理可能存在的旧实例
	if old, err := GetRedisPool(name); err == nil {
		_ = old.ClosePool()
	}

	inst1 := NewRedis(name, testAddr, testPasswd)
	inst2 := NewRedis(name, testAddr, testPasswd)
	defer func() {
		_ = inst1.ClosePool()
	}()

	if inst1 != inst2 {
		t.Error("两次 NewRedis 应返回同一实例（LoadOrStore 语义）")
	}
}

// TestNewRedis_ConcurrentSafe 验证并发调用 NewRedis 不会 panic 或泄漏
func TestNewRedis_ConcurrentSafe(t *testing.T) {
	name := "test_concurrent_safe"
	if old, err := GetRedisPool(name); err == nil {
		_ = old.ClosePool()
	}

	var wg sync.WaitGroup
	instances := make([]*RdPool, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			instances[idx] = NewRedis(name, testAddr, testPasswd)
		}(i)
	}
	wg.Wait()

	// 所有实例应该是同一个
	for i := 1; i < 10; i++ {
		if instances[i] != instances[0] {
			t.Errorf("并发 NewRedis 返回了不同实例: instances[0]=%p, instances[%d]=%p", instances[0], i, instances[i])
		}
	}
	_ = instances[0].ClosePool()
}

// ========== 2. ClosePool 正确删除 ==========

// TestClosePool_DeleteFromMap 验证 ClosePool 能正确从 map 中删除
func TestClosePool_DeleteFromMap(t *testing.T) {
	name := "test_close_pool"
	inst := NewRedis(name, testAddr, testPasswd)

	err := inst.ClosePool()
	if err != nil {
		t.Errorf("ClosePool 失败: %v", err)
	}

	// 关闭后应该无法通过 GetRedisPool 获取
	_, err = GetRedisPool(name)
	if err == nil {
		t.Error("ClosePool 后仍能获取到实例，说明 Name 字段未正确初始化或 Delete 未生效")
	}
}

// ========== 3. Do 通用接口 — 可变参数展开 ==========

// TestDo_VarArgs 验证 Do 方法正确展开可变参数
func TestDo_VarArgs(t *testing.T) {
	inst := getTestRedis(t)

	// 使用 Do 执行 SET 命令
	_, err := inst.Do("SET", "test_do_key", "test_do_value")
	if err != nil {
		t.Fatalf("Do SET 失败: %v", err)
	}

	// 使用 Do 执行 GET 命令验证
	ret, err := inst.Do("GET", "test_do_key")
	if err != nil {
		t.Fatalf("Do GET 失败: %v", err)
	}

	val, ok := ret.([]byte)
	if !ok {
		t.Fatalf("Do GET 返回类型错误: %T", ret)
	}
	if string(val) != "test_do_value" {
		t.Errorf("期望 %q, 实际 %q", "test_do_value", string(val))
	}

	// 清理
	_, _ = inst.Del("test_do_key")
}

// ========== 4. HSet 返回实际值 ==========

// TestHSet_ReturnActualValue 验证 HSet 返回实际值而非硬编码 0
func TestHSet_ReturnActualValue(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_hset_return"

	// 清理
	_, _ = inst.Del(key)

	// 第一次 HSet，应返回 1（新建字段）
	ret, err := inst.HSet(key, "field1", "value1")
	if err != nil {
		t.Fatalf("HSet 失败: %v", err)
	}
	if ret != 1 {
		t.Errorf("首次 HSet 期望返回 1（新建字段），实际返回 %d", ret)
	}

	// 第二次 HSet 同一字段，应返回 0（更新已有字段）
	ret, err = inst.HSet(key, "field1", "value2")
	if err != nil {
		t.Fatalf("HSet 失败: %v", err)
	}
	if ret != 0 {
		t.Errorf("更新 HSet 期望返回 0（更新字段），实际返回 %d", ret)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 5. HMGet 参数展开 ==========

// TestHMGet_MultipleFields 验证 HMGet 正确展开可变参数
func TestHMGet_MultipleFields(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_hmget"

	// 清理并准备数据
	_, _ = inst.Del(key)
	_, _ = inst.HSet(key, "f1", "v1")
	_, _ = inst.HSet(key, "f2", "v2")
	_, _ = inst.HSet(key, "f3", "v3")

	// 一次获取多个字段
	ret, err := inst.HMGet(key, "f1", "f2", "f3")
	if err != nil {
		t.Fatalf("HMGet 失败: %v", err)
	}
	if len(ret) != 3 {
		t.Fatalf("HMGet 期望返回 3 个值，实际返回 %d 个", len(ret))
	}
	expected := []string{"v1", "v2", "v3"}
	for i, v := range ret {
		if v != expected[i] {
			t.Errorf("HMGet[%d] 期望 %q, 实际 %q", i, expected[i], v)
		}
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 6. HMSet 参数展开 ==========

// TestHMSet_MapArgs 验证 HMSet 正确展开 map 参数
func TestHMSet_MapArgs(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_hmset"

	// 清理
	_, _ = inst.Del(key)

	kv := map[string]string{
		"field1": "hello",
		"field2": "world",
	}
	ret, err := inst.HMSet(key, kv)
	if err != nil {
		t.Fatalf("HMSet 失败: %v", err)
	}
	if ret != "OK" {
		t.Errorf("HMSet 期望返回 OK, 实际返回 %q", ret)
	}

	// 验证数据
	v1, err := inst.HGet(key, "field1")
	if err != nil {
		t.Fatalf("HGet field1 失败: %v", err)
	}
	if v1 != "hello" {
		t.Errorf("期望 field1=%q, 实际 %q", "hello", v1)
	}

	v2, err := inst.HGet(key, "field2")
	if err != nil {
		t.Fatalf("HGet field2 失败: %v", err)
	}
	if v2 != "world" {
		t.Errorf("期望 field2=%q, 实际 %q", "world", v2)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 7. HDel 参数展开 ==========

// TestHDel_MultipleFields 验证 HDel 正确展开可变参数
func TestHDel_MultipleFields(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_hdel"

	// 清理并准备数据
	_, _ = inst.Del(key)
	_, _ = inst.HSet(key, "f1", "v1")
	_, _ = inst.HSet(key, "f2", "v2")
	_, _ = inst.HSet(key, "f3", "v3")

	// 一次删除多个字段
	ret, err := inst.HDel(key, "f1", "f2")
	if err != nil {
		t.Fatalf("HDel 失败: %v", err)
	}
	if ret != 2 {
		t.Errorf("HDel 期望删除 2 个字段，实际删除 %d 个", ret)
	}

	// 验证 f3 仍然存在
	v, err := inst.HGet(key, "f3")
	if err != nil {
		t.Fatalf("HGet f3 失败: %v", err)
	}
	if v != "v3" {
		t.Errorf("期望 f3=%q, 实际 %q", "v3", v)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 8. HGetAll 命令名大写 ==========

// TestHGetAll_Correct 验证 HGetAll 正常工作
func TestHGetAll_Correct(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_hgetall"

	// 清理并准备数据
	_, _ = inst.Del(key)
	_, _ = inst.HSet(key, "name", "alice")
	_, _ = inst.HSet(key, "age", "30")

	ret, err := inst.HGetAll(key)
	if err != nil {
		t.Fatalf("HGetAll 失败: %v", err)
	}
	if len(ret) != 2 {
		t.Fatalf("HGetAll 期望返回 2 个字段，实际返回 %d 个", len(ret))
	}
	if ret["name"] != "alice" {
		t.Errorf("期望 name=%q, 实际 %q", "alice", ret["name"])
	}
	if ret["age"] != "30" {
		t.Errorf("期望 age=%q, 实际 %q", "30", ret["age"])
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 9. LPush 一次性 push ==========

// TestLPush_MultipleValues 验证 LPush 一次性 push 多个值
func TestLPush_MultipleValues(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_lpush_multi"

	// 清理
	_, _ = inst.Del(key)

	// 一次 push 多个值
	ret, err := inst.LPush(key, "a", "b", "c")
	if err != nil {
		t.Fatalf("LPush 失败: %v", err)
	}
	if ret != 3 {
		t.Errorf("LPush 期望返回列表长度 3，实际返回 %d", ret)
	}

	// 验证列表内容（LPUSH 是头部插入，顺序为 c, b, a）
	list, err := inst.LRange(key, 0, -1)
	if err != nil {
		t.Fatalf("LRange 失败: %v", err)
	}
	expected := []string{"c", "b", "a"}
	if len(list) != len(expected) {
		t.Fatalf("期望列表长度 %d, 实际 %d", len(expected), len(list))
	}
	for i, v := range list {
		if v != expected[i] {
			t.Errorf("LRange[%d] 期望 %q, 实际 %q", i, expected[i], v)
		}
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 10. LPushX 修正为 LPUSHX 命令 ==========

// TestLPushX_OnlyPushWhenKeyExists 验证 LPushX 仅在 key 存在时 push
func TestLPushX_OnlyPushWhenKeyExists(t *testing.T) {
	inst := getTestRedis(t)
	keyExists := "test_lpushx_exists"
	keyNotExists := "test_lpushx_not_exists"

	// 清理
	_, _ = inst.Del(keyExists)
	_, _ = inst.Del(keyNotExists)

	// 先创建一个列表
	_, _ = inst.LPush(keyExists, "initial")

	// LPushX 到已存在的 key，应该成功
	ret, err := inst.LPushX(keyExists, "new_value")
	if err != nil {
		t.Fatalf("LPushX 到已存在 key 失败: %v", err)
	}
	if ret != 2 {
		t.Errorf("LPushX 期望返回列表长度 2，实际返回 %d", ret)
	}

	// LPushX 到不存在的 key，应该返回 0（不插入）
	ret, err = inst.LPushX(keyNotExists, "should_not_exist")
	if err != nil {
		t.Fatalf("LPushX 到不存在 key 失败: %v", err)
	}
	if ret != 0 {
		t.Errorf("LPushX 到不存在 key 期望返回 0，实际返回 %d", ret)
	}

	// 验证不存在的 key 确实没有数据
	llen, err := inst.LLen(keyNotExists)
	if err != nil {
		t.Fatalf("LLen 失败: %v", err)
	}
	if llen != 0 {
		t.Errorf("不存在的 key 期望长度 0，实际 %d", llen)
	}

	// 清理
	_, _ = inst.Del(keyExists)
}

// ========== 11. LSet 参数顺序修正 + 返回类型 ==========

// TestLSet_CorrectParamOrder 验证 LSet 参数顺序正确（key, index, value）
func TestLSet_CorrectParamOrder(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_lset"

	// 清理并准备数据
	_, _ = inst.Del(key)
	_, _ = inst.LPush(key, "one", "two", "three")
	// 列表内容: three, two, one

	// LSet 修改 index=0 的元素
	ret, err := inst.LSet(key, 0, "MODIFIED")
	if err != nil {
		t.Fatalf("LSet 失败: %v", err)
	}
	if ret != "OK" {
		t.Errorf("LSet 期望返回 %q, 实际返回 %q", "OK", ret)
	}

	// 验证修改结果
	val, err := inst.LIndex(key, 0)
	if err != nil {
		t.Fatalf("LIndex 失败: %v", err)
	}
	if val != "MODIFIED" {
		t.Errorf("LSet 后 index=0 期望 %q, 实际 %q", "MODIFIED", val)
	}

	// 测试负数索引
	ret, err = inst.LSet(key, -1, "LAST")
	if err != nil {
		t.Fatalf("LSet 负数索引失败: %v", err)
	}
	if ret != "OK" {
		t.Errorf("LSet 期望返回 %q, 实际返回 %q", "OK", ret)
	}

	val, err = inst.LIndex(key, -1)
	if err != nil {
		t.Fatalf("LIndex 失败: %v", err)
	}
	if val != "LAST" {
		t.Errorf("LSet 后 index=-1 期望 %q, 实际 %q", "LAST", val)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 12. MGet 参数展开 ==========

// TestMGet_MultipleKeys 验证 MGet 正确展开可变参数
func TestMGet_MultipleKeys(t *testing.T) {
	inst := getTestRedis(t)

	// 准备数据
	_, _ = inst.Set("test_mget_k1", "v1")
	_, _ = inst.Set("test_mget_k2", "v2")
	_, _ = inst.Set("test_mget_k3", "v3")

	ret, err := inst.MGet("test_mget_k1", "test_mget_k2", "test_mget_k3")
	if err != nil {
		t.Fatalf("MGet 失败: %v", err)
	}
	if len(ret) != 3 {
		t.Fatalf("MGet 期望返回 3 个值，实际返回 %d 个", len(ret))
	}
	expected := []string{"v1", "v2", "v3"}
	for i, v := range ret {
		if v != expected[i] {
			t.Errorf("MGet[%d] 期望 %q, 实际 %q", i, expected[i], v)
		}
	}

	// 清理
	_, _ = inst.Del("test_mget_k1")
	_, _ = inst.Del("test_mget_k2")
	_, _ = inst.Del("test_mget_k3")
}

// ========== 13. MSet 参数展开 ==========

// TestMSet_MultiplePairs 验证 MSet 正确展开可变参数
func TestMSet_MultiplePairs(t *testing.T) {
	inst := getTestRedis(t)

	ret, err := inst.MSet("test_mset_k1", "hello", "test_mset_k2", "world")
	if err != nil {
		t.Fatalf("MSet 失败: %v", err)
	}
	if ret != "OK" {
		t.Errorf("MSet 期望返回 OK, 实际返回 %q", ret)
	}

	// 验证数据
	v1, _ := inst.Get("test_mset_k1")
	v2, _ := inst.Get("test_mset_k2")
	if v1 != "hello" {
		t.Errorf("期望 test_mset_k1=%q, 实际 %q", "hello", v1)
	}
	if v2 != "world" {
		t.Errorf("期望 test_mset_k2=%q, 实际 %q", "world", v2)
	}

	// 清理
	_, _ = inst.Del("test_mset_k1")
	_, _ = inst.Del("test_mset_k2")
}

// ========== 14. SetEX 参数顺序修正 ==========

// TestSetEX_CorrectParamOrder 验证 SetEX 参数顺序正确（key, seconds, value）
func TestSetEX_CorrectParamOrder(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_setex"

	// 清理
	_, _ = inst.Del(key)

	ret, err := inst.SetEX(key, "hello_setex", 60)
	if err != nil {
		t.Fatalf("SetEX 失败: %v", err)
	}
	if ret != "OK" {
		t.Errorf("SetEX 期望返回 OK, 实际返回 %q", ret)
	}

	// 验证值
	val, err := inst.Get(key)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "hello_setex" {
		t.Errorf("期望值 %q, 实际 %q", "hello_setex", val)
	}

	// 验证 TTL
	ttl, err := inst.Ttl(key)
	if err != nil {
		t.Fatalf("Ttl 失败: %v", err)
	}
	if ttl <= 0 || ttl > 60 {
		t.Errorf("SetEX TTL 期望在 (0, 60] 范围内，实际 %d", ttl)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 15. SetNX 改用 SET ... EX ... NX ==========

// TestSetNX_WithExpire 验证 SetNX 带过期时间的原子设置
func TestSetNX_WithExpire(t *testing.T) {
	inst := getTestRedis(t)
	key := "test_setnx"

	// 清理
	_, _ = inst.Del(key)

	// 第一次 SetNX，key 不存在，应该成功
	ret, err := inst.SetNX(key, "first_value", 60)
	if err != nil {
		t.Fatalf("SetNX 失败: %v", err)
	}
	if ret != 1 {
		t.Errorf("首次 SetNX 期望返回 1（设置成功），实际返回 %d", ret)
	}

	// 验证值
	val, err := inst.Get(key)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "first_value" {
		t.Errorf("期望值 %q, 实际 %q", "first_value", val)
	}

	// 验证 TTL（应该有过期时间）
	ttl, err := inst.Ttl(key)
	if err != nil {
		t.Fatalf("Ttl 失败: %v", err)
	}
	if ttl <= 0 || ttl > 60 {
		t.Errorf("SetNX TTL 期望在 (0, 60] 范围内，实际 %d", ttl)
	}

	// 第二次 SetNX，key 已存在，应该失败
	ret, err = inst.SetNX(key, "second_value", 60)
	if err != nil {
		t.Fatalf("SetNX 失败: %v", err)
	}
	if ret != 0 {
		t.Errorf("重复 SetNX 期望返回 0（key 已存在），实际返回 %d", ret)
	}

	// 验证值未被覆盖
	val, err = inst.Get(key)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "first_value" {
		t.Errorf("SetNX 不应覆盖已有值，期望 %q, 实际 %q", "first_value", val)
	}

	// 清理
	_, _ = inst.Del(key)
}

// ========== 16. 连接池健康检查和超时配置 ==========

// TestPoolConfig_Ping 验证连接池配置正确（通过 Ping 间接验证）
func TestPoolConfig_Ping(t *testing.T) {
	inst := getTestRedis(t)

	ret, err := inst.Ping()
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
	if ret != "PONG" {
		t.Errorf("Ping 期望返回 PONG, 实际返回 %q", ret)
	}
}

// TestPoolConfig_ConnectTimeout 验证连接超时配置生效（连接不可达地址）
func TestPoolConfig_ConnectTimeout(t *testing.T) {
	name := "test_timeout"
	// 使用一个不可达的地址
	inst := NewRedis(name, "192.0.2.1:6379", "")
	defer func() {
		_ = inst.ClosePool()
	}()

	start := time.Now()
	_, err := inst.Ping()
	elapsed := time.Since(start)

	if err == nil {
		t.Error("连接不可达地址应该返回错误")
	}
	// 超时应该在 5 秒左右（DialConnectTimeout 设置为 5s）
	if elapsed > 10*time.Second {
		t.Errorf("连接超时时间过长: %v, 期望在 5s 左右", elapsed)
	}
	t.Logf("连接超时耗时: %v", elapsed)
}

// ========== 17. GetRedisPool / GetRedisByName ==========

// TestGetRedisPool_Success 验证正常获取
func TestGetRedisPool_Success(t *testing.T) {
	name := "test_get_pool"
	inst := NewRedis(name, testAddr, testPasswd)
	defer func() {
		_ = inst.ClosePool()
	}()

	got, err := GetRedisPool(name)
	if err != nil {
		t.Fatalf("GetRedisPool 失败: %v", err)
	}
	if got != inst {
		t.Error("GetRedisPool 返回的实例与 NewRedis 创建的不一致")
	}
}

// TestGetRedisPool_NotFound 验证获取不存在的实例
func TestGetRedisPool_NotFound(t *testing.T) {
	_, err := GetRedisPool("non_existent_pool_12345")
	if err == nil {
		t.Error("获取不存在的 pool 应该返回错误")
	}
}

// TestGetRedisByName_NotFound 验证获取不存在的实例返回 nil
func TestGetRedisByName_NotFound(t *testing.T) {
	inst := GetRedisByName("non_existent_pool_12345")
	if inst != nil {
		t.Error("获取不存在的 pool 应该返回 nil")
	}
}

// ========== 原有测试保留 ==========

func TestRedis(t *testing.T) {
	rediInst := NewRedis("redis1", testAddr, testPasswd)

	// get the redis instance by name
	rediInst = GetRedisByName("redis1")

	set, err := rediInst.Set("chairou", "111")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log(set)

	inc, err := rediInst.Increment("chairou")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("inc:", inc)

	get, err := rediInst.Get("chairou")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log(get)

	expire, err := rediInst.Expired("chairou", 500)
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	if expire == 0 {
		t.Log("expire doesn't work")
	}
	t.Log(expire)

	ttl, err := rediInst.Ttl("chairou")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("ttl:", ttl)

	hset, err := rediInst.HSet("chairou_hset", "test", "222")
	if err != nil {
		t.Log("hset", err, hset)
		t.Error(err)
		return
	}
	t.Log("hset: ", hset)
	hget, err := rediInst.HGet("chairou_hset", "test")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log(hget)

	hgetall, err := rediInst.HGetAll("chairou_hset")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("hgetall:", hgetall)

	hdel, err := rediInst.HDel("chair_hset", "test")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("hdel:", hdel)

	del, err := rediInst.Del("chairou")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("del:", del)

	hsetex, err := rediInst.HSetEX("chair", "redisPool2", "111", 10)
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log("hsetex:", hsetex)

	_, err = rediInst.LPush("list1", "1")
	_, err = rediInst.LPush("list1", "abc")
	llen, err := rediInst.LLen("list1")
	if err != nil {
		t.Error(err)
	}
	t.Log("llen:", llen)

	llist, err := rediInst.LRange("list1", 0, -1)
	t.Log("llist:", llist)

	lrem, err := rediInst.LRem("list1", 0, "abc")
	if err != nil {
		t.Error(err)
	}
	t.Log("lrem:", lrem)

	llist2, err := rediInst.LRange("list1", 0, -1)
	t.Log("llist:", llist2)

	lpop, err := rediInst.LPop("list1")
	t.Log("lpop:", lpop)
	_, err = rediInst.Del("list1")

	lpushx, err := rediInst.LPushX("asd", "qqqq")
	t.Log("lpushx:", lpushx)

	redisPool2, err := GetRedisPool("redis1")
	if err != nil {
		t.Error("GetRedisPool err:", err)
	}
	_, err = redisPool2.Set("redisPool2", "bbb")
	if err != nil {
		t.Error("Set err:", err)
		return
	}

	str, err := redisPool2.Get("ffffffffffffffffffff")
	if err.Error() != "redigo: nil returned" {
		t.Error("Get err:", err)
	}
	if str != "" {
		t.Error("Get err:", err)
	}

	str = GetRedisByName("redis1").SilenceGet("ffffffffffffffffffff")
	if str != "" {
		t.Error("SilenceGet err:", "ffffffffffffffffffffffff")
	}

	set, err = rediInst.Set("chairou", "")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log(set)
	get, err = rediInst.Get("chairou0000000000000000")
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}
	t.Log(get)
}

func TestList(t *testing.T) {
	list := make([]int, 10)
	list = append(list, 9)
	t.Log(list)
}
