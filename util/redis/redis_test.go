package redis

import (
	"testing"
)

// go test -v redis_test.go redis.go

func TestRedis(t *testing.T) {
	rediInst := NewRedis("redis1", "127.0.0.1:6379", "chairou")

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

}
