package cache

import (
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/runtime"

	"github.com/gomodule/redigo/redis"
)

type CacheRedis struct {
	cli *redis.Pool

	prefix string
}

func (p *CacheRedis) Clean() error {
	conn := p.cli.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", p.prefix+"*"))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	_, err = conn.Do("DEL", args...)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheRedis) Incr(key string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("INCR", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Decr(key string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("DECR", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) IncrBy(key string, value int64) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("INCRBY", p.prefix+key, value))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) IncrByFloat(key string, increment float64) (float64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Float64(conn.Do("INCRBYFLOAT", p.prefix+key, increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) DecrBy(key string, value int64) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("DECRBY", p.prefix+key, value))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Get(key string) (string, error) {
	log.Debugf("get %s", key)

	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("GET", p.prefix+key))
	if err != nil {
		if err == redis.ErrNil {
			return "", ErrNotFound
		}
		log.Errorf("err:%v", err)
		return "", err
	}

	return val, nil
}

func (p *CacheRedis) Exists(keys ...string) (bool, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		args = append(args, p.prefix+key)
	}

	count, err := redis.Int64(conn.Do("EXISTS", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return count > 0, nil
}

func (p *CacheRedis) SetNx(key string, value interface{}) (bool, error) {
	log.Debugf("set nx %s", key)

	conn := p.cli.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("SETNX", p.prefix+key, candy.ToString(value)))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Expire(key string, timeout time.Duration) (bool, error) {
	conn := p.cli.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXPIRE", p.prefix+key, int64(timeout.Seconds())))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Ttl(key string) (time.Duration, error) {
	conn := p.cli.Get()
	defer conn.Close()

	ttl, err := redis.Int64(conn.Do("TTL", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return time.Duration(ttl) * time.Second, nil
}

func (p *CacheRedis) Set(key string, value interface{}) (err error) {
	log.Debugf("set %s", key)

	conn := p.cli.Get()
	defer conn.Close()

	_, err = conn.Do("SET", p.prefix+key, candy.ToString(value))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) SetEx(key string, value interface{}, timeout time.Duration) error {
	log.Debugf("set ex %s", key)

	conn := p.cli.Get()
	defer conn.Close()

	_, err := conn.Do("SETEX", p.prefix+key, int64(timeout.Seconds()), candy.ToString(value))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	log.Debugf("set nx ex %s", key)

	ok, err := p.SetNx(key, value)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	if ok {
		_, err = p.Expire(key, timeout)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return ok, nil
}

func (p *CacheRedis) Del(keys ...string) (err error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		args = append(args, p.prefix+key)
	}

	_, err = conn.Do("DEL", args...)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) HSet(key string, field string, value interface{}) (bool, error) {
	conn := p.cli.Get()
	defer conn.Close()

	result, err := redis.Int64(conn.Do("HSET", p.prefix+key, field, candy.ToString(value)))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return result > 0, nil
}

func (p *CacheRedis) HGet(key, field string) (string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("HGET", p.prefix+key, field))
	if err != nil {
		if err == redis.ErrNil {
			return "", ErrNotFound
		}
		log.Errorf("err:%v", err)
		return "", err
	}

	return val, nil
}

func (p *CacheRedis) HGetAll(key string) (map[string]string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.StringMap(conn.Do("HGETALL", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HKeys(key string) ([]string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("HKEYS", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HVals(key string) ([]string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("HVALS", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HDel(key string, fields ...string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, p.prefix+key)
	for _, field := range fields {
		args = append(args, field)
	}

	val, err := redis.Int64(conn.Do("HDEL", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SAdd(key string, members ...string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, p.prefix+key)
	for _, member := range members {
		args = append(args, member)
	}

	val, err := redis.Int64(conn.Do("SADD", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SMembers(key string) ([]string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("SMEMBERS", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) SRem(key string, members ...string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, p.prefix+key)
	for _, member := range members {
		args = append(args, member)
	}

	val, err := redis.Int64(conn.Do("SREM", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SRandMember(key string, count ...int64) ([]string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0)
	args = append(args, p.prefix+key)
	if len(count) > 0 {
		args = append(args, count[0])
	} else {
		args = append(args, 1)
	}

	val, err := redis.Strings(conn.Do("SRANDMEMBER", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) SPop(key string) (string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("SPOP", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (p *CacheRedis) HIncr(key string, subKey string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HINCRBY", p.prefix+key, subKey, 1))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HIncrBy(key string, field string, increment int64) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HINCRBY", p.prefix+key, field, increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Float64(conn.Do("HINCRBYFLOAT", p.prefix+key, field, increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HDecr(key string, field string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HINCRBY", p.prefix+key, field, -1))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HDecrBy(key string, field string, increment int64) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HINCRBY", p.prefix+key, field, -increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HExists(key, field string) (bool, error) {
	conn := p.cli.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("HEXISTS", p.prefix+key, field))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (p *CacheRedis) SisMember(key, field string) (bool, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, 2)
	args = append(args, p.prefix+key, field)

	val, err := redis.Bool(conn.Do("SISMEMBER", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return val, nil
}

func (p *CacheRedis) Close() error {
	err := p.cli.Close()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) Ping() error {
	conn := p.cli.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) Publish(channel string, message interface{}) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("PUBLISH", p.prefix+channel, candy.ToString(message)))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	conn := p.cli.Get()
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}

	prefixedChannels := make([]interface{}, len(channels))
	for i, ch := range channels {
		prefixedChannels[i] = p.prefix + ch
	}

	err := psc.Subscribe(prefixedChannels...)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	logic := func(v redis.Message) {
		defer runtime.CachePanicWithHandle(func(err interface{}) {
			log.Errorf("PANIC:%v", err)
		})

		channel := v.Channel
		if len(p.prefix) > 0 && len(channel) > len(p.prefix) {
			channel = channel[len(p.prefix):]
		}

		err := handler(channel, v.Data)
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			logic(v)
		case redis.Subscription:
			log.Debugf("subscription: %s %s %d", v.Kind, v.Channel, v.Count)
		case error:
			log.Errorf("err:%v", v)
			return v
		}
	}
}

func (p *CacheRedis) XAdd(stream string, values map[string]interface{}) (string, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(values)*2+2)
	args = append(args, p.prefix+stream, "*")
	for k, v := range values {
		args = append(args, k, candy.ToString(v))
	}

	val, err := redis.String(conn.Do("XADD", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (p *CacheRedis) XLen(stream string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("XLEN", p.prefix+stream))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, 4)
	args = append(args, p.prefix+stream, start, stop)
	if len(count) > 0 && count[0] > 0 {
		args = append(args, "COUNT", count[0])
	}

	reply, err := redis.Values(conn.Do("XRANGE", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(reply))
	for _, item := range reply {
		entry, err := redis.Values(item, nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		if len(entry) != 2 {
			continue
		}

		id, err := redis.String(entry[0], nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		fields, err := redis.StringMap(entry[1], nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		entryMap := make(map[string]interface{})
		entryMap["id"] = id
		for k, v := range fields {
			entryMap[k] = v
		}
		result = append(result, entryMap)
	}

	return result, nil
}

func (p *CacheRedis) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, 4)
	args = append(args, p.prefix+stream, start, stop)
	if len(count) > 0 && count[0] > 0 {
		args = append(args, "COUNT", count[0])
	}

	reply, err := redis.Values(conn.Do("XREVRANGE", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(reply))
	for _, item := range reply {
		entry, err := redis.Values(item, nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		if len(entry) != 2 {
			continue
		}

		id, err := redis.String(entry[0], nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		fields, err := redis.StringMap(entry[1], nil)
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}

		entryMap := make(map[string]interface{})
		entryMap["id"] = id
		for k, v := range fields {
			entryMap[k] = v
		}
		result = append(result, entryMap)
	}

	return result, nil
}

func (p *CacheRedis) XDel(stream string, ids ...string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, p.prefix+stream)
	for _, id := range ids {
		args = append(args, id)
	}

	val, err := redis.Int64(conn.Do("XDEL", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) XTrim(stream string, maxLen int64) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("XTRIM", p.prefix+stream, "MAXLEN", maxLen))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XGroupCreate 创建消费者组
// stream: Stream 键名
// group: 消费者组名称
// start: 起始消息 ID，可以是具体 ID、"0"（从头开始）或 "$"（从最新消息开始）
// MKSTREAM 选项会在 stream 不存在时自动创建
func (p *CacheRedis) XGroupCreate(stream, group, start string) error {
	conn := p.cli.Get()
	defer conn.Close()

	_, err := conn.Do("XGROUP", "CREATE", p.prefix+stream, group, start, "MKSTREAM")
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XGroupDestroy 销毁消费者组
// stream: Stream 键名
// group: 消费者组名称
func (p *CacheRedis) XGroupDestroy(stream, group string) error {
	conn := p.cli.Get()
	defer conn.Close()

	_, err := conn.Do("XGROUP", "DESTROY", p.prefix+stream, group)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XGroupSetID 设置消费者组的起始消息 ID
// stream: Stream 键名
// group: 消费者组名称
// id: 新的起始消息 ID
func (p *CacheRedis) XGroupSetID(stream, group, id string) error {
	conn := p.cli.Get()
	defer conn.Close()

	_, err := conn.Do("XGROUP", "SETID", p.prefix+stream, group, id)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XReadGroup 使用消费者组持续消费 Stream 消息（回调模式，参考 Subscribe 设计）
// handler: 消息处理回调函数
//   - stream: Stream 键名（已去除 prefix）
//   - id: 消息 ID
//   - body: 消息体的原始字节数组（直接从 Redis 获取，无需序列化）
//   - 返回 error 时会退出消费循环
//
// group: 消费者组名称
// consumer: 消费者名称
// stream: Stream 键名
//
// 特点：
//   - 使用 BLOCK 60000 阻塞等待新消息（60秒超时，有效防止 CPU 空转）
//   - 使用 ">" 特殊 ID 只读取未投递给其他消费者的新消息
//   - 自动处理 panic 恢复（参考 Subscribe 的设计）
//   - 自动处理 prefix 的添加和移除
//   - handler 返回 error 时优雅退出
//   - 消息需要手动调用 XAck 确认
//   - XREADGROUP 本身已经阻塞，无需额外 sleep
//   - 假设消息使用单个字段存储数据（如 "data"），直接返回该字段的值
func (p *CacheRedis) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	conn := p.cli.Get()
	defer conn.Close()

	// 消息处理逻辑，包含 panic 恢复机制
	logic := func(streamName string, id string, body []byte) {
		defer runtime.CachePanicWithHandle(func(err interface{}) {
			log.Errorf("PANIC:%v", err)
		})

		// 移除 prefix，返回原始 stream 名称
		originalStream := streamName
		if len(p.prefix) > 0 && len(streamName) > len(p.prefix) {
			originalStream = streamName[len(p.prefix):]
		}

		err := handler(originalStream, id, body)
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	}

	// 无限循环，持续消费消息
	for {
		// XREADGROUP GROUP <group> <consumer> BLOCK <timeout> STREAMS <stream> <id>
		// BLOCK 30000 表示阻塞等待 30000ms（30秒）
		// - 当有新消息时，立即返回（延迟小于 30 秒）
		// - 当无消息时，阻塞 60 秒后返回空结果（CPU 几乎不占用）
		// - 相比 BLOCK 0 的优势：可以定期检查连接状态、支持优雅退出
		// ">" 表示只读取未投递给其他消费者的新消息
		reply, err := redis.Values(conn.Do("XREADGROUP", "GROUP", group, consumer, "BLOCK", 60000, "STREAMS", p.prefix+stream, ">"))
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// 如果没有消息（超时返回），直接继续下一轮
		// XREADGROUP 本身已经阻塞了 30 秒，不需要额外 sleep
		if len(reply) == 0 {
			continue
		}

		// 解析返回的 stream 数据
		// 格式: [[stream_name, [[id1, [field1, value1, ...]], [id2, [field2, value2, ...]]]]]
		for _, streamData := range reply {
			streamInfo, err := redis.Values(streamData, nil)
			if err != nil {
				log.Errorf("err:%v", err)
				continue
			}

			if len(streamInfo) != 2 {
				continue
			}

			// 解析 stream 名称
			streamName, err := redis.String(streamInfo[0], nil)
			if err != nil {
				log.Errorf("err:%v", err)
				continue
			}

			// 解析消息列表
			messages, err := redis.Values(streamInfo[1], nil)
			if err != nil {
				log.Errorf("err:%v", err)
				continue
			}

			// 处理每条消息
			for _, msg := range messages {
				entry, err := redis.Values(msg, nil)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				if len(entry) != 2 {
					continue
				}

				// 解析消息 ID
				id, err := redis.String(entry[0], nil)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				// 解析消息字段列表（field-value pairs）
				// entry[1] 是 [field1, value1, field2, value2, ...] 格式
				fieldValues, err := redis.Values(entry[1], nil)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				// 假设消息使用单个字段 "data" 存储实际数据
				// 如果字段数量不是 2（1对 field-value），则跳过
				if len(fieldValues) != 2 {
					log.Warnf("unexpected field count: %d, expected 2 (single field-value pair)", len(fieldValues))
					continue
				}

				// 获取字段值（索引 1 是 value）
				body, err := redis.Bytes(fieldValues[1], nil)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				// 调用回调函数处理消息
				logic(streamName, id, body)
			}
		}
	}
}

// XAck 确认消费者组中的消息
// stream: Stream 键名
// group: 消费者组名称
// ids: 要确认的消息 ID 列表
// 返回成功确认的消息数量
func (p *CacheRedis) XAck(stream, group string, ids ...string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	args := make([]interface{}, 0, len(ids)+2)
	args = append(args, p.prefix+stream, group)
	for _, id := range ids {
		args = append(args, id)
	}

	val, err := redis.Int64(conn.Do("XACK", args...))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XPending 查询消费者组中待处理消息的数量
// stream: Stream 键名
// group: 消费者组名称
// 返回待处理消息的总数量
func (p *CacheRedis) XPending(stream, group string) (int64, error) {
	conn := p.cli.Get()
	defer conn.Close()

	reply, err := redis.Values(conn.Do("XPENDING", p.prefix+stream, group))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	if len(reply) == 0 {
		return 0, nil
	}

	count, err := redis.Int64(reply[0], nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return count, nil
}

func (p *CacheRedis) Redis() *redis.Pool {
	return p.cli
}

func NewRedis(address string, opts ...redis.DialOption) (Cache, error) {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, opts...)
		},
		MaxIdle:     1000,
		MaxActive:   1000,
		IdleTimeout: time.Second * 5,
		Wait:        true,
	}

	p := &CacheRedis{
		cli: pool,
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	err := p.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return newBaseCache(p), nil
}
