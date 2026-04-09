package redis

import (
	"errors"
	"log"
	"sync"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

var redisMap sync.Map

type RdPool struct {
	pool   *redigo.Pool
	Name   string
	addr   string
	passwd string
}

// NewRedis 生成新的redis实例并放入Pool中
func NewRedis(name string, addr string, passwd string) *RdPool {
	newInst := &RdPool{Name: name, addr: addr, passwd: passwd}
	newInst.newRedisPool(addr, passwd)
	inst, loaded := redisMap.LoadOrStore(name, newInst)
	if loaded {
		// 已存在，关闭刚创建的连接池
		_ = newInst.pool.Close()
		return inst.(*RdPool)
	}
	return newInst
}

// GetRedisPool 每次用前先获得redis pool的实例
func GetRedisPool(name string) (*RdPool, error) {
	inst, ok := redisMap.Load(name)
	if ok {
		return inst.(*RdPool), nil
	} else {
		return nil, errors.New("get redis pool from syncMap err.")
	}
}

// GetRedisByName 不处理错误, 连写方式
func GetRedisByName(name string) *RdPool {
	inst, ok := redisMap.Load(name)
	if ok {
		return inst.(*RdPool)
	} else {
		log.Println("GetRedisByName err:", name)
		return nil
	}
}

func (c *RdPool) newRedisPool(addr string, passwd string) {
	setPasswd := redigo.DialPassword(passwd)
	c.pool = &redigo.Pool{
		MaxIdle:     5,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redigo.Conn, error) {
			conn, err := redigo.Dial("tcp", addr,
				setPasswd,
				redigo.DialConnectTimeout(5*time.Second),
				redigo.DialReadTimeout(3*time.Second),
				redigo.DialWriteTimeout(3*time.Second),
			)
			if err != nil {
				return nil, errors.New("failed to connect to Redis: " + err.Error())
			}
			return conn, nil
		},
		MaxActive: 100,
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// Get
// redis> GET mykey
// "Hello"
// redis>
func (c *RdPool) Get(key string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	str, err := redigo.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	return str, nil
}

// SilenceGet 不会返回错误, 只返回空
func (c *RdPool) SilenceGet(key string) string {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	str, err := redigo.String(conn.Do("GET", key))
	if err != nil {
		return ""
	}
	return str
}

// HGet
// redis> HGET myhash field1
// "foo"
func (c *RdPool) HGet(key string, subKey string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	str, err := redigo.String(conn.Do("HGET", key, subKey))
	if err != nil {
		return "", err
	}
	return str, nil
}

// Set
// redis> SET mykey "Hello"
// "OK"
func (c *RdPool) Set(key string, val string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	str, err := redigo.String(conn.Do("SET", key, val))
	if err != nil {
		return "", err
	}
	return str, nil
}

// HSet
// redis> HSET myhash field1 "foo"
// (integer) 1
func (c *RdPool) HSet(key string, subKey string, val string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("HSET", key, subKey, val))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// Del
// redis> SET key1 "Hello"
// "OK"
// redis> SET key2 "World"
// "OK"
// redis> DEL key1 key2 key3
// (integer) 2
func (c *RdPool) Del(key string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	str, err := redigo.Int64(conn.Do("DEL", key))
	if err != nil {
		return str, err
	}
	return str, nil
}

// Do 通用接口
func (c *RdPool) Do(commandName string, args ...interface{}) (interface{}, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := conn.Do(commandName, args...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// ClosePool 关闭连接池
func (c *RdPool) ClosePool() error {
	err := c.pool.Close()
	redisMap.Delete(c.Name)
	if err != nil {
		return err
	}
	return nil
}

// Expired
// redis> SET mykey "Hello"
// "OK"
// redis> EXPIRE mykey 10
// (integer) 1
// redis> TTL mykey
// (integer) 10
func (c *RdPool) Expired(key string, seconds int) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("EXPIRE", key, seconds))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (c *RdPool) Ttl(key string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("TTL", key))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// HMGet
// redis> HSET myhash field1 "Hello"
// (integer) 1
// redis> HSET myhash field2 "World"
// (integer) 1
// redis> HMGET myhash field1 field2 nofield
// 1) "Hello"
// 2) "World"
// 3) (nil)
func (c *RdPool) HMGet(key string, values ...string) ([]string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.Add(key).AddFlat(values)
	ret, err := redigo.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// HMSet
// redis> HMSET myhash field1 "Hello" field2 "World"
// "OK"
// redis> HGET myhash field1
// "Hello"
// redis> HGET myhash field2
// "World"
func (c *RdPool) HMSet(key string, kv map[string]string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.Add(key).AddFlat(kv)
	ret, err := redigo.String(conn.Do("HMSET", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// HDel
// redis> HSET myhash field1 "foo"
// (integer) 1
// redis> HDEL myhash field1
// (integer) 1
func (c *RdPool) HDel(key string, fields ...string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.Add(key).AddFlat(fields)
	ret, err := redigo.Int64(conn.Do("HDEL", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// HGetAll
// redis> HSET myhash field1 "Hello"
// (integer) 1
// redis> HSET myhash field2 "World"
// (integer) 1
// redis> HGETALL myhash
// 1) "field1"
// 2) "Hello"
// 3) "field2"
// 4) "World"
func (c *RdPool) HGetAll(key string) (map[string]string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// HSetEX
// 127.0.0.1:6379> hset chair aaa 111
// (integer) 1
// 127.0.0.1:6379> expire chair 10
// (integer) 1
// 127.0.0.1:6379> ttl chair
// (integer) 8
func (c *RdPool) HSetEX(key, field string, value interface{}, expire int) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("HSet", key, field, value))
	if err != nil {
		return ret, err
	}
	ret, err = redigo.Int64(conn.Do("EXPIRE", key, expire))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// Increment
// redis> SET mykey "10"
// "OK"
// redis> INCR mykey
// (integer) 11
// redis> GET mykey
// "11"
func (c *RdPool) Increment(key string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("INCR", key))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LIndex
// redis> LPUSH mylist "World"
// (integer) 1
// redis> LPUSH mylist "Hello"
// (integer) 2
// redis> LINDEX mylist 0
// "Hello"
// redis> LINDEX mylist -1
// "World"
// redis> LINDEX mylist 3
// (nil)
func (c *RdPool) LIndex(key string, index int) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("LIndex", key, index))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LLen
// redis> LPUSH mylist "World"
// (integer) 1
// redis> LPUSH mylist "Hello"
// (integer) 2
// redis> LLEN mylist
// (integer) 2
// redis>
func (c *RdPool) LLen(key string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("LLen", key))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LPop
// redis> RPUSH mylist "one" "two" "three" "four" "five"
// (integer) 5
// redis> LPOP mylist
// "one"
// redis> LPOP mylist 2
// 1) "two"
// 2) "three"
// redis> LRANGE mylist 0 -1
// 1) "four"
// 2) "five"
func (c *RdPool) LPop(key string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("LPop", key))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LPush
// redis> LPUSH mylist "world"
// (integer) 1
// redis> LPUSH mylist "hello"
// (integer) 2
// redis> LRANGE mylist 0 -1
// 1) "hello"
// 2) "world"
func (c *RdPool) LPush(key string, values ...interface{}) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.Add(key).AddFlat(values)
	ret, err := redigo.Int64(conn.Do("LPUSH", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LPushX
// redis> LPUSH mylist "World"
// (integer) 1
// redis> LPUSHX mylist "Hello"
// (integer) 2
// redis> LPUSHX myotherlist "Hello"
// (integer) 0
// redis> LRANGE mylist 0 -1
// 1) "Hello"
// 2) "World"
// redis> LRANGE myotherlist 0 -1
// (empty array)
func (c *RdPool) LPushX(key string, values ...interface{}) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.Add(key).AddFlat(values)
	ret, err := redigo.Int64(conn.Do("LPUSHX", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LRem
// redis> RPUSH mylist "hello"
// (integer) 1
// redis> RPUSH mylist "hello"
// (integer) 2
// redis> RPUSH mylist "foo"
// (integer) 3
// redis> RPUSH mylist "hello"
// (integer) 4
// redis> LREM mylist -2 "hello"
// (integer) 2
// redis> LRANGE mylist 0 -1
// 1) "hello"
// 2) "foo"
func (c *RdPool) LRem(key string, count int, value string) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Int64(conn.Do("LREM", key, count, value))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LSet
// redis> RPUSH mylist "one"
// (integer) 1
// redis> RPUSH mylist "two"
// (integer) 2
// redis> RPUSH mylist "three"
// (integer) 3
// redis> LSET mylist 0 "four"
// "OK"
// redis> LSET mylist -2 "five"
// "OK"
// redis> LRANGE mylist 0 -1
// 1) "four"
// 2) "five"
// 3) "three"
func (c *RdPool) LSet(key string, index int, value string) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("LSET", key, index, value))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// LTrim
// redis> RPUSH mylist "one"
// (integer) 1
// redis> RPUSH mylist "two"
// (integer) 2
// redis> RPUSH mylist "three"
// (integer) 3
// redis> LTRIM mylist 1 -1
// "OK"
// redis> LRANGE mylist 0 -1
// 1) "two"
// 2) "three"
func (c *RdPool) LTrim(key string, start, stop int) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("LTrim", key, start, stop))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (c *RdPool) MGet(keys ...string) ([]string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	args := redigo.Args{}.AddFlat(keys)
	ret, err := redigo.Strings(conn.Do("MGET", args...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// MSet redis> MSET key1 "Hello" key2 "World"
func (c *RdPool) MSet(pairs ...interface{}) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("MSET", pairs...))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// Ping
// redis> PING
// "PONG"
func (c *RdPool) Ping() (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("Ping"))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// SetEx
// redis> SETEX mykey 10 "Hello"
// "OK"
// redis> TTL mykey
// (integer) 10
// redis> GET mykey
// "Hello"
// redis>
func (c *RdPool) SetEX(key string, value interface{}, expire int) (string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.String(conn.Do("SETEX", key, expire, value))
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// SetNX
// redis> SETNX mykey "Hello"
// (integer) 1
// redis> SETNX mykey "World"
// (integer) 0
// redis> GET mykey
// "Hello"
func (c *RdPool) SetNX(key string, value interface{}, expire int) (int64, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	// 使用 SET key value EX seconds NX 实现带过期时间的原子设置
	// 返回 "OK" 表示设置成功，nil 表示 key 已存在
	reply, err := conn.Do("SET", key, value, "EX", expire, "NX")
	if err != nil {
		return 0, err
	}
	if reply == nil {
		return 0, nil // key 已存在，未设置
	}
	return 1, nil // 设置成功
}

func (c *RdPool) LRange(key string, start, stop int64) ([]string, error) {
	conn := c.pool.Get()
	defer func(conn redigo.Conn) {
		_ = conn.Close()
	}(conn)
	ret, err := redigo.Strings(conn.Do("LRange", key, start, stop))
	if err != nil {
		return ret, err
	}
	return ret, nil
}
