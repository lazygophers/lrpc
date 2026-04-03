package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/runtime"
	"github.com/redis/go-redis/v9"
)

type CacheRedis struct {
	cli       *redis.Client
	prefix    string
	ctx       context.Context
	miniRedis *miniredis.Miniredis // mock Redis 实例，仅在 Mock 模式下使用
}

func (p *CacheRedis) Clean() error {
	ctx := p.ctx
	var cursor uint64

	pattern := p.prefix + "*"
	for {
		keys, nextCursor, err := p.cli.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if len(keys) > 0 {
			err = p.cli.Del(ctx, keys...).Err()
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (p *CacheRedis) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheRedis) Incr(key string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.Incr(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Decr(key string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.Decr(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) IncrBy(key string, value int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.IncrBy(ctx, p.prefix+key, value).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) IncrByFloat(key string, increment float64) (float64, error) {
	ctx := p.ctx
	val, err := p.cli.IncrByFloat(ctx, p.prefix+key, increment).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) DecrBy(key string, value int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.DecrBy(ctx, p.prefix+key, value).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Get(key string) (string, error) {
	log.Debugf("get %s", key)

	ctx := p.ctx
	val, err := p.cli.Get(ctx, p.prefix+key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound
		}
		log.Errorf("err:%v", err)
		return "", err
	}

	return val, nil
}

func (p *CacheRedis) Exists(keys ...string) (bool, error) {
	ctx := p.ctx
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = p.prefix + key
	}

	count, err := p.cli.Exists(ctx, prefixedKeys...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return count > 0, nil
}

func (p *CacheRedis) SetNx(key string, value interface{}) (bool, error) {
	log.Debugf("set nx %s", key)

	ctx := p.ctx
	ok, err := p.cli.SetNX(ctx, p.prefix+key, candy.ToString(value), 0).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Expire(key string, timeout time.Duration) (bool, error) {
	ctx := p.ctx
	ok, err := p.cli.Expire(ctx, p.prefix+key, timeout).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Ttl(key string) (time.Duration, error) {
	ctx := p.ctx
	ttl, err := p.cli.TTL(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return ttl, nil
}

func (p *CacheRedis) Set(key string, value interface{}) error {
	log.Debugf("set %s", key)

	ctx := p.ctx
	err := p.cli.Set(ctx, p.prefix+key, candy.ToString(value), 0).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) SetEx(key string, value interface{}, timeout time.Duration) error {
	log.Debugf("set ex %s", key)

	ctx := p.ctx
	err := p.cli.Set(ctx, p.prefix+key, candy.ToString(value), timeout).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	log.Debugf("set nx ex %s", key)

	ctx := p.ctx
	ok, err := p.cli.SetNX(ctx, p.prefix+key, candy.ToString(value), timeout).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Del(keys ...string) error {
	ctx := p.ctx
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = p.prefix + key
	}

	err := p.cli.Del(ctx, prefixedKeys...).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) HSet(key string, field string, value interface{}) (bool, error) {
	ctx := p.ctx
	result, err := p.cli.HSet(ctx, p.prefix+key, field, candy.ToString(value)).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return result > 0, nil
}

func (p *CacheRedis) HGet(key, field string) (string, error) {
	ctx := p.ctx
	val, err := p.cli.HGet(ctx, p.prefix+key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound
		}
		log.Errorf("err:%v", err)
		return "", err
	}

	return val, nil
}

func (p *CacheRedis) HGetAll(key string) (map[string]string, error) {
	ctx := p.ctx
	val, err := p.cli.HGetAll(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HKeys(key string) ([]string, error) {
	ctx := p.ctx
	val, err := p.cli.HKeys(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HVals(key string) ([]string, error) {
	ctx := p.ctx
	val, err := p.cli.HVals(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) HDel(key string, fields ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.HDel(ctx, p.prefix+key, fields...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SAdd(key string, members ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.SAdd(ctx, p.prefix+key, members).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SMembers(key string) ([]string, error) {
	ctx := p.ctx
	val, err := p.cli.SMembers(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) SRem(key string, members ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.SRem(ctx, p.prefix+key, members).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) SRandMember(key string, count ...int64) ([]string, error) {
	ctx := p.ctx
	c := int64(1)
	if len(count) > 0 {
		c = count[0]
	}

	val, err := p.cli.SRandMemberN(ctx, p.prefix+key, c).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

func (p *CacheRedis) SPop(key string) (string, error) {
	ctx := p.ctx
	val, err := p.cli.SPop(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (p *CacheRedis) HIncr(key string, subKey string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.HIncrBy(ctx, p.prefix+key, subKey, 1).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HIncrBy(key string, field string, increment int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.HIncrBy(ctx, p.prefix+key, field, increment).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	ctx := p.ctx
	val, err := p.cli.HIncrByFloat(ctx, p.prefix+key, field, increment).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HDecr(key string, field string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.HIncrBy(ctx, p.prefix+key, field, -1).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HDecrBy(key string, field string, increment int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.HIncrBy(ctx, p.prefix+key, field, -increment).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) HExists(key, field string) (bool, error) {
	ctx := p.ctx
	ok, err := p.cli.HExists(ctx, p.prefix+key, field).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (p *CacheRedis) SisMember(key, field string) (bool, error) {
	ctx := p.ctx
	val, err := p.cli.SIsMember(ctx, p.prefix+key, field).Result()
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

	// 如果使用的是 miniredis，关闭它
	if p.miniRedis != nil {
		p.miniRedis.Close()
	}

	return nil
}

func (p *CacheRedis) Ping() error {
	ctx := p.ctx
	_, err := p.cli.Ping(ctx).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheRedis) Publish(channel string, message interface{}) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.Publish(ctx, p.prefix+channel, candy.ToString(message)).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	ctx := p.ctx
	prefixedChannels := make([]string, len(channels))
	for i, ch := range channels {
		prefixedChannels[i] = p.prefix + ch
	}

	pubsub := p.cli.Subscribe(ctx, prefixedChannels...)
	defer pubsub.Close()

	// 处理消息
	logic := func(msg *redis.Message) {
		defer runtime.CachePanicWithHandle(func(err interface{}) {
			log.Errorf("PANIC:%v", err)
		})

		channel := msg.Channel
		if len(p.prefix) > 0 && len(channel) > len(p.prefix) {
			channel = channel[len(p.prefix):]
		}

		err := handler(channel, []byte(msg.Payload))
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	}

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		logic(msg)
	}
}

// XAdd 添加消息到 Stream
func (p *CacheRedis) XAdd(stream string, values map[string]interface{}) (string, error) {
	ctx := p.ctx

	// 构造 Redis Stream 参数
	args := &redis.XAddArgs{
		Stream: p.prefix + stream,
		ID:     "*",
	}

	// 转换 values 为字符串
	fields := make(map[string]interface{})
	for k, v := range values {
		fields[k] = candy.ToString(v)
	}
	args.Values = fields

	id, err := p.cli.XAdd(ctx, args).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return id, nil
}

// XLen 返回 Stream 长度
func (p *CacheRedis) XLen(stream string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.XLen(ctx, p.prefix+stream).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XRange 返回 Stream 中的消息范围
func (p *CacheRedis) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	ctx := p.ctx

	var result []map[string]interface{}

	if len(count) > 0 && count[0] > 0 {
		// 使用 COUNT 限制
		msgs, err := p.cli.XRangeN(ctx, p.prefix+stream, start, stop, count[0]).Result()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		for _, msg := range msgs {
			entry := make(map[string]interface{})
			entry["id"] = msg.ID
			for k, v := range msg.Values {
				entry[k] = v
			}
			result = append(result, entry)
		}
	} else {
		msgs, err := p.cli.XRange(ctx, p.prefix+stream, start, stop).Result()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		for _, msg := range msgs {
			entry := make(map[string]interface{})
			entry["id"] = msg.ID
			for k, v := range msg.Values {
				entry[k] = v
			}
			result = append(result, entry)
		}
	}

	return result, nil
}

// XRevRange 返回 Stream 中的反向消息范围
func (p *CacheRedis) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	ctx := p.ctx

	var result []map[string]interface{}

	if len(count) > 0 && count[0] > 0 {
		msgs, err := p.cli.XRevRangeN(ctx, p.prefix+stream, start, stop, count[0]).Result()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		for _, msg := range msgs {
			entry := make(map[string]interface{})
			entry["id"] = msg.ID
			for k, v := range msg.Values {
				entry[k] = v
			}
			result = append(result, entry)
		}
	} else {
		msgs, err := p.cli.XRevRange(ctx, p.prefix+stream, start, stop).Result()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		for _, msg := range msgs {
			entry := make(map[string]interface{})
			entry["id"] = msg.ID
			for k, v := range msg.Values {
				entry[k] = v
			}
			result = append(result, entry)
		}
	}

	return result, nil
}

// XDel 删除 Stream 中的消息
func (p *CacheRedis) XDel(stream string, ids ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.XDel(ctx, p.prefix+stream, ids...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XTrim 裁剪 Stream 到指定长度
func (p *CacheRedis) XTrim(stream string, maxLen int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.XTrimMaxLen(ctx, p.prefix+stream, maxLen).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XGroupCreate 创建消费者组
func (p *CacheRedis) XGroupCreate(stream, group, start string) error {
	ctx := p.ctx
	err := p.cli.XGroupCreate(ctx, p.prefix+stream, group, start).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XGroupDestroy 销毁消费者组
func (p *CacheRedis) XGroupDestroy(stream, group string) error {
	ctx := p.ctx
	err := p.cli.XGroupDestroy(ctx, p.prefix+stream, group).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XGroupSetID 设置消费者组的起始消息 ID
func (p *CacheRedis) XGroupSetID(stream, group, id string) error {
	ctx := p.ctx
	err := p.cli.XGroupSetID(ctx, p.prefix+stream, group, id).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

// XReadGroup 使用消费者组持续消费 Stream 消息（回调模式）
func (p *CacheRedis) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	ctx := p.ctx

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
		// XREADGROUP BLOCK 60000 表示阻塞等待 60000ms（60秒）
		// ">" 表示只读取未投递给其他消费者的新消息
		streams, err := p.cli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{p.prefix + stream, ">"},
			Count:    1,
			Block:    60 * time.Second,
		}).Result()

		if err != nil && err != redis.Nil {
			log.Errorf("err:%v", err)
			return err
		}

		// 如果没有消息（超时返回），直接继续下一轮
		if len(streams) == 0 {
			continue
		}

		// 解析返回的 stream 数据
		for _, streamData := range streams {
			// 处理每条消息
			for _, msg := range streamData.Messages {
				// 假设消息使用单个字段 "data" 存储实际数据
				// 如果字段数量不是 1，则跳过
				if len(msg.Values) != 1 {
					log.Warnf("unexpected field count: %d, expected 1 (single field-value pair)", len(msg.Values))
					continue
				}

				// 获取字段值
				var body []byte
				for _, v := range msg.Values {
					body = []byte(fmt.Sprint(v))
					break
				}

				// 调用回调函数处理消息
				logic(streamData.Stream, msg.ID, body)
			}
		}
	}
}

// XAck 确认消费者组中的消息
func (p *CacheRedis) XAck(stream, group string, ids ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.XAck(ctx, p.prefix+stream, group, ids...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// XPending 查询消费者组中待处理消息的数量
func (p *CacheRedis) XPending(stream, group string) (int64, error) {
	ctx := p.ctx
	pending, err := p.cli.XPending(ctx, p.prefix+stream, group).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return pending.Count, nil
}

// ZAdd 添加成员到有序集合
func (p *CacheRedis) ZAdd(key string, members ...interface{}) (int64, error) {
	ctx := p.ctx

	// 将可变参数转换为redis.Z结构
	zMembers := make([]redis.Z, 0, len(members)/2)
	for i := 0; i < len(members); i += 2 {
		if i+1 >= len(members) {
			break
		}
		score, ok := members[i].(float64)
		if !ok {
			// 尝试从其他数值类型转换
			if intScore, ok := members[i].(int); ok {
				score = float64(intScore)
			} else if int64Score, ok := members[i].(int64); ok {
				score = float64(int64Score)
			} else {
				continue
			}
		}
		member := candy.ToString(members[i+1])
		zMembers = append(zMembers, redis.Z{Score: score, Member: member})
	}

	val, err := p.cli.ZAdd(ctx, p.prefix+key, zMembers...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZScore 获取成员分数
func (p *CacheRedis) ZScore(key, member string) (float64, error) {
	ctx := p.ctx
	val, err := p.cli.ZScore(ctx, p.prefix+key, member).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNotFound
		}
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZRange 按索引范围获取成员（升序）
func (p *CacheRedis) ZRange(key string, start, stop int64) ([]string, error) {
	ctx := p.ctx
	val, err := p.cli.ZRange(ctx, p.prefix+key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

// ZRangeByScore 按分数范围获取成员
func (p *CacheRedis) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	ctx := p.ctx
	opt := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}
	val, err := p.cli.ZRangeByScore(ctx, p.prefix+key, opt).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

// ZRem 删除成员
func (p *CacheRedis) ZRem(key string, members ...string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZRem(ctx, p.prefix+key, members).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZCard 获取集合成员数
func (p *CacheRedis) ZCard(key string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZCard(ctx, p.prefix+key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZCount 统计分数范围内成员数
func (p *CacheRedis) ZCount(key, min, max string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZCount(ctx, p.prefix+key, min, max).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZIncrBy 增加成员分数
func (p *CacheRedis) ZIncrBy(key string, increment float64, member string) (float64, error) {
	ctx := p.ctx
	val, err := p.cli.ZIncrBy(ctx, p.prefix+key, increment, member).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZRank 获取成员排名（升序，从0开始）
func (p *CacheRedis) ZRank(key, member string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZRank(ctx, p.prefix+key, member).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, ErrNotFound
		}
		log.Errorf("err:%v", err)
		return -1, err
	}
	return val, nil
}

// ZRevRange 按索引范围获取成员（降序）
func (p *CacheRedis) ZRevRange(key string, start, stop int64) ([]string, error) {
	ctx := p.ctx
	val, err := p.cli.ZRevRange(ctx, p.prefix+key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

// ZRevRank 获取成员排名（降序，从0开始）
func (p *CacheRedis) ZRevRank(key, member string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZRevRank(ctx, p.prefix+key, member).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, ErrNotFound
		}
		log.Errorf("err:%v", err)
		return -1, err
	}
	return val, nil
}

// ZRangeWithScores 按索引范围获取成员和分数
func (p *CacheRedis) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	ctx := p.ctx
	val, err := p.cli.ZRangeWithScores(ctx, p.prefix+key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := make([]Z, len(val))
	for i, z := range val {
		result[i] = Z{
			Member: z.Member.(string),
			Score:  z.Score,
		}
	}
	return result, nil
}

// ZRevRangeByScore 按分数范围获取成员（降序）
func (p *CacheRedis) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	ctx := p.ctx
	opt := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}
	val, err := p.cli.ZRevRangeByScore(ctx, p.prefix+key, opt).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return val, nil
}

// ZRemRangeByRank 按排名范围删除成员
func (p *CacheRedis) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZRemRangeByRank(ctx, p.prefix+key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZRemRangeByScore 按分数范围删除成员
func (p *CacheRedis) ZRemRangeByScore(key, min, max string) (int64, error) {
	ctx := p.ctx
	val, err := p.cli.ZRemRangeByScore(ctx, p.prefix+key, min, max).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZUnionStore 并集存储
func (p *CacheRedis) ZUnionStore(destination string, keys ...string) (int64, error) {
	ctx := p.ctx

	// 添加prefix
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = p.prefix + key
	}

	store := &redis.ZStore{
		Keys: prefixedKeys,
	}

	val, err := p.cli.ZUnionStore(ctx, p.prefix+destination, store).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

// ZInterStore 交集存储
func (p *CacheRedis) ZInterStore(destination string, keys ...string) (int64, error) {
	ctx := p.ctx

	// 添加prefix
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = p.prefix + key
	}

	store := &redis.ZStore{
		Keys: prefixedKeys,
	}

	val, err := p.cli.ZInterStore(ctx, p.prefix+destination, store).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheRedis) Client() any {
	return p.cli
}

// NewRedis 创建 Redis 缓存实例
// address: Redis 服务器地址（格式：host:port）
// opts: redis.DialOption 选项（已废弃，保留用于兼容性）
//
// Deprecated: 使用 NewRedisWithClient 或 NewRedisWithConfig 替代
func NewRedis(address string, opts ...interface{}) (Cache, error) {
	// 解析旧版本选项（兼容性）
	db := 0
	password := ""
	for _, opt := range opts {
		// 尝试从选项中提取配置
		if m, ok := opt.(map[string]interface{}); ok {
			if v, ok := m["db"]; ok {
				db = int(v.(int64))
			}
			if v, ok := m["password"]; ok {
				password = v.(string)
			}
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     password,
		DB:           db,
		PoolSize:     1, // Default to 1 to reduce memory usage
		MinIdleConns: 0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	return NewRedisWithClient(client, "")
}

// NewRedisWithConfig 使用配置创建 Redis 缓存
func NewRedisWithConfig(config *Config) (Cache, error) {
	if config == nil {
		config = &Config{}
	}
	config.apply()

	// 如果启用 Mock，使用 miniredis
	if config.Mock {
		mr, err := miniredis.Run()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		return NewRedisWithMiniRedis(client, "", mr)
	}

	client := redis.NewClient(&redis.Options{
		Addr:            config.Address,
		Password:        config.Password,
		DB:              config.Db,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		MaxIdleConns:    config.MaxIdleConns,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxLifetime: config.ConnMaxLifetime,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
	})

	return NewRedisWithClient(client, "")
}

// NewRedisWithClient 使用已有的 redis.Client 创建缓存
func NewRedisWithClient(client *redis.Client, prefix string) (Cache, error) {
	return NewRedisWithMiniRedis(client, prefix, nil)
}

// NewRedisWithMiniRedis 使用已有的 redis.Client 和可选的 miniredis 创建缓存
func NewRedisWithMiniRedis(client *redis.Client, prefix string, mr *miniredis.Miniredis) (Cache, error) {
	p := &CacheRedis{
		cli:       client,
		prefix:    prefix,
		ctx:       context.Background(),
		miniRedis: mr,
	}

	// 测试连接
	err := p.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return newBaseCache(p), nil
}
