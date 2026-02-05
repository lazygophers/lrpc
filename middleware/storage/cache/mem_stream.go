package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/lazygophers/log"
	"gorm.io/gorm/utils"
)

func (p *CacheMem) XAdd(stream string, values map[string]interface{}) (string, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		p.streams = make(map[string]*Stream)
	}

	s, exists := p.streams[stream]
	if !exists {
		s = &Stream{
			Messages: make([]*StreamMessage, 0),
			Groups:   make(map[string]*ConsumerGroup),
		}
		p.streams[stream] = s
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成新的消息 ID (使用时间戳-序列号格式)
	id := fmt.Sprintf("%d-%d", time.Now().UnixMilli(), atomic.AddInt64(&p.streamID, 1))

	msg := &StreamMessage{
		ID:        id,
		Values:    values,
		CreatedAt: time.Now(),
		Acked:     false,
	}

	s.Messages = append(s.Messages, msg)
	return id, nil
}

func (p *CacheMem) XLen(stream string) (int64, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.Messages)), nil
}

func (p *CacheMem) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return []map[string]interface{}{}, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return []map[string]interface{}{}, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	startIdx := 0
	endIdx := len(s.Messages)

	// 解析 start 参数
	if start != "-" && start != "0" {
		for i, msg := range s.Messages {
			if msg.ID >= start {
				startIdx = i
				break
			}
		}
	}

	// 解析 stop 参数
	if stop != "+" {
		for i := len(s.Messages) - 1; i >= 0; i-- {
			if s.Messages[i].ID <= stop {
				endIdx = i + 1
				break
			}
		}
	}

	if startIdx >= endIdx {
		return []map[string]interface{}{}, nil
	}

	// 应用 count 限制
	maxCount := int64(-1)
	if len(count) > 0 && count[0] > 0 {
		maxCount = count[0]
	}

	result := make([]map[string]interface{}, 0)
	for i := startIdx; i < endIdx && (maxCount == -1 || int64(len(result)) < maxCount); i++ {
		result = append(result, map[string]interface{}{
			"id":     s.Messages[i].ID,
			"values": s.Messages[i].Values,
		})
	}

	return result, nil
}

func (p *CacheMem) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return []map[string]interface{}{}, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return []map[string]interface{}{}, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	startIdx := len(s.Messages) - 1
	endIdx := 0

	// 解析 start 参数
	if start != "+" {
		for i := len(s.Messages) - 1; i >= 0; i-- {
			if s.Messages[i].ID <= start {
				startIdx = i
				break
			}
		}
	}

	// 解析 stop 参数
	if stop != "-" && stop != "0" {
		for i, msg := range s.Messages {
			if msg.ID >= stop {
				endIdx = i
				break
			}
		}
	}

	if startIdx < endIdx {
		return []map[string]interface{}{}, nil
	}

	// 应用 count 限制
	maxCount := int64(-1)
	if len(count) > 0 && count[0] > 0 {
		maxCount = count[0]
	}

	result := make([]map[string]interface{}, 0)
	for i := startIdx; i >= endIdx && (maxCount == -1 || int64(len(result)) < maxCount); i-- {
		result = append(result, map[string]interface{}{
			"id":     s.Messages[i].ID,
			"values": s.Messages[i].Values,
		})
	}

	return result, nil
}

func (p *CacheMem) XDel(stream string, ids ...string) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	deletedCount := int64(0)
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	newMessages := make([]*StreamMessage, 0)
	for _, msg := range s.Messages {
		if idMap[msg.ID] {
			deletedCount++
		} else {
			newMessages = append(newMessages, msg)
		}
	}

	s.Messages = newMessages

	// 同时从所有消费者组的待处理消息中删除
	for _, group := range s.Groups {
		for _, id := range ids {
			delete(group.Pending, id)
		}
	}

	return deletedCount, nil
}

func (p *CacheMem) XTrim(stream string, maxLen int64) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldLen := len(s.Messages)
	if int64(oldLen) <= maxLen {
		return 0, nil
	}

	// 保留最新的 maxLen 条消息
	s.Messages = s.Messages[oldLen-int(maxLen):]
	return int64(oldLen - len(s.Messages)), nil
}

func (p *CacheMem) XGroupCreate(stream, group, start string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		p.streams = make(map[string]*Stream)
	}

	s, exists := p.streams[stream]
	if !exists {
		s = &Stream{
			Messages: make([]*StreamMessage, 0),
			Groups:   make(map[string]*ConsumerGroup),
		}
		p.streams[stream] = s
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Groups[group]; exists {
		return nil // 组已存在，直接返回
	}

	// 确定起始 ID
	lastID := "0"
	if start == "$" {
		// 从当前最后一条消息开始
		if len(s.Messages) > 0 {
			lastID = s.Messages[len(s.Messages)-1].ID
		}
	} else if start != "0" {
		lastID = start
	}

	s.Groups[group] = &ConsumerGroup{
		Name:      group,
		LastID:    lastID,
		Pending:   make(map[string]*StreamMessage),
		Consumers: make(map[string]string),
	}

	return nil
}

func (p *CacheMem) XGroupDestroy(stream, group string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Groups, group)
	return nil
}

func (p *CacheMem) XGroupSetID(stream, group, id string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return errors.New("stream not found")
	}

	s, exists := p.streams[stream]
	if !exists {
		return errors.New("stream not found")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.Groups[group]
	if !exists {
		return errors.New("group not found")
	}

	g.LastID = id
	return nil
}

func (p *CacheMem) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	p.streamsMu.Lock()

	if p.streams == nil {
		p.streamsMu.Unlock()
		return errors.New("stream not found")
	}

	s, exists := p.streams[stream]
	if !exists {
		p.streamsMu.Unlock()
		return errors.New("stream not found")
	}

	p.streamsMu.Unlock()

	// 持续消费消息
	for {
		s.mu.Lock()

		g, exists := s.Groups[group]
		if !exists {
			s.mu.Unlock()
			return errors.New("group not found")
		}

		// 查找新消息（ID > LastID）
		var newMsgs []*StreamMessage
		for _, msg := range s.Messages {
			if msg.ID > g.LastID && !msg.Acked {
				newMsgs = append(newMsgs, msg)
			}
		}

		// 如果有新消息，处理它们
		if len(newMsgs) > 0 {
			for _, msg := range newMsgs {
				// 将消息添加到待处理列表
				g.Pending[msg.ID] = msg
				g.Consumers[consumer] = msg.ID
				g.LastID = msg.ID

				s.mu.Unlock()

				// 提取消息体（假设使用单个字段 "data"）
				var body []byte
				if len(msg.Values) == 1 {
					for _, v := range msg.Values {
						body = []byte(utils.ToString(v))
						break
					}
				} else {
					// 如果有多个字段，序列化为 JSON
					jsonData, err := json.Marshal(msg.Values)
					if err != nil {
						log.Errorf("err:%v", err)
						body = []byte("{}")
					} else {
						body = jsonData
					}
				}

				// 调用处理函数
				err := handler(stream, msg.ID, body)
				if err != nil {
					log.Errorf("err:%v", err)
				}

				s.mu.Lock()
			}
		} else {
			s.mu.Unlock()
		}

		s.mu.Unlock()

		// 短暂休眠避免 CPU 占用过高
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *CacheMem) XAck(stream, group string, ids ...string) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.Groups[group]
	if !exists {
		return 0, nil
	}

	ackedCount := int64(0)
	for _, id := range ids {
		if msg, exists := g.Pending[id]; exists {
			msg.Acked = true
			delete(g.Pending, id)
			ackedCount++
		}
	}

	return ackedCount, nil
}

func (p *CacheMem) XPending(stream, group string) (int64, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	g, exists := s.Groups[group]
	if !exists {
		return 0, nil
	}

	return int64(len(g.Pending)), nil
}
