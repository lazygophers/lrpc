package cache

import (
	"fmt"
	"sort"
	"sync"
)

// ZSet 有序集合数据结构
type ZSet struct {
	mu      sync.RWMutex
	members map[string]float64 // member -> score映射
}

// ZSetMember 有序集合成员（用于排序）
type ZSetMember struct {
	Member string
	Score  float64
}

// newZSet 创建新的ZSet
func newZSet() *ZSet {
	return &ZSet{
		members: make(map[string]float64),
	}
}

// add 添加成员
func (z *ZSet) add(member string, score float64) bool {
	_, exists := z.members[member]
	z.members[member] = score
	return !exists // 返回是否是新成员
}

// remove 删除成员
func (z *ZSet) remove(member string) bool {
	_, exists := z.members[member]
	if exists {
		delete(z.members, member)
	}
	return exists
}

// score 获取成员分数
func (z *ZSet) score(member string) (float64, bool) {
	score, exists := z.members[member]
	return score, exists
}

// card 获取成员数量
func (z *ZSet) card() int {
	return len(z.members)
}

// getSortedMembers 获取排序后的成员列表
func (z *ZSet) getSortedMembers(ascending bool) []ZSetMember {
	members := make([]ZSetMember, 0, len(z.members))
	for member, score := range z.members {
		members = append(members, ZSetMember{
			Member: member,
			Score:  score,
		})
	}

	sort.Slice(members, func(i, j int) bool {
		if members[i].Score != members[j].Score {
			if ascending {
				return members[i].Score < members[j].Score
			}
			return members[i].Score > members[j].Score
		}
		// 分数相同时按字典序排序
		if ascending {
			return members[i].Member < members[j].Member
		}
		return members[i].Member > members[j].Member
	})

	return members
}

// rank 获取成员排名（升序）
func (z *ZSet) rank(member string, ascending bool) (int64, bool) {
	if _, exists := z.members[member]; !exists {
		return -1, false
	}

	sorted := z.getSortedMembers(ascending)
	for i, m := range sorted {
		if m.Member == member {
			return int64(i), true
		}
	}
	return -1, false
}

// CacheMem ZSet 方法实现

// ZAdd 添加成员到有序集合
func (p *CacheMem) ZAdd(key string, members ...interface{}) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}
	if p.zsets[key] == nil {
		p.zsets[key] = newZSet()
	}

	zset := p.zsets[key]
	zset.mu.Lock()
	defer zset.mu.Unlock()

	// 解析members（score, member pairs）
	var count int64
	for i := 0; i < len(members); i += 2 {
		if i+1 >= len(members) {
			break
		}

		var score float64
		switch v := members[i].(type) {
		case float64:
			score = v
		case float32:
			score = float64(v)
		case int:
			score = float64(v)
		case int64:
			score = float64(v)
		case int32:
			score = float64(v)
		default:
			continue
		}

		member := ""
		switch v := members[i+1].(type) {
		case string:
			member = v
		default:
			continue
		}

		if zset.add(member, score) {
			count++
		}
	}

	return count, nil
}

// ZScore 获取成员分数
func (p *CacheMem) ZScore(key, member string) (float64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	score, exists := zset.score(member)
	if !exists {
		return 0, ErrNotFound
	}

	return score, nil
}

// ZRange 按索引范围获取成员（升序）
func (p *CacheMem) ZRange(key string, start, stop int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(true)
	return p.sliceRange(sorted, start, stop), nil
}

// ZRangeByScore 按分数范围获取成员
func (p *CacheMem) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	sorted := zset.getSortedMembers(true)
	result := make([]string, 0)

	for _, m := range sorted {
		if (minInclusive && m.Score >= minScore) || (!minInclusive && m.Score > minScore) {
			if (maxInclusive && m.Score <= maxScore) || (!maxInclusive && m.Score < maxScore) {
				result = append(result, m.Member)
			}
		}
	}

	// 应用offset和count
	if offset > 0 {
		if offset >= int64(len(result)) {
			return []string{}, nil
		}
		result = result[offset:]
	}
	if count > 0 && count < int64(len(result)) {
		result = result[:count]
	}

	return result, nil
}

// ZRem 删除成员
func (p *CacheMem) ZRem(key string, members ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	var count int64
	for _, member := range members {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZCard 获取集合成员数
func (p *CacheMem) ZCard(key string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	return int64(zset.card()), nil
}

// ZCount 统计分数范围内成员数
func (p *CacheMem) ZCount(key, min, max string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	var count int64
	for _, score := range zset.members {
		if (minInclusive && score >= minScore) || (!minInclusive && score > minScore) {
			if (maxInclusive && score <= maxScore) || (!maxInclusive && score < maxScore) {
				count++
			}
		}
	}

	return count, nil
}

// ZIncrBy 增加成员分数
func (p *CacheMem) ZIncrBy(key string, increment float64, member string) (float64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}
	if p.zsets[key] == nil {
		p.zsets[key] = newZSet()
	}

	zset := p.zsets[key]
	zset.mu.Lock()
	defer zset.mu.Unlock()

	score, exists := zset.score(member)
	if !exists {
		score = 0
	}

	newScore := score + increment
	zset.add(member, newScore)

	return newScore, nil
}

// ZRank 获取成员排名（升序，从0开始）
func (p *CacheMem) ZRank(key, member string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return -1, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	rank, exists := zset.rank(member, true)
	if !exists {
		return -1, ErrNotFound
	}

	return rank, nil
}

// ZRevRange 按索引范围获取成员（降序）
func (p *CacheMem) ZRevRange(key string, start, stop int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(false)
	return p.sliceRange(sorted, start, stop), nil
}

// ZRevRank 获取成员排名（降序，从0开始）
func (p *CacheMem) ZRevRank(key, member string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return -1, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	rank, exists := zset.rank(member, false)
	if !exists {
		return -1, ErrNotFound
	}

	return rank, nil
}

// ZRangeWithScores 按索引范围获取成员和分数
func (p *CacheMem) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []Z{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(true)
	members := p.sliceRangeWithScore(sorted, start, stop)

	result := make([]Z, len(members))
	for i, m := range members {
		result[i] = Z{
			Member: m.Member,
			Score:  m.Score,
		}
	}

	return result, nil
}

// ZRevRangeByScore 按分数范围获取成员（降序）
func (p *CacheMem) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	sorted := zset.getSortedMembers(false)
	result := make([]string, 0)

	for _, m := range sorted {
		if (maxInclusive && m.Score <= maxScore) || (!maxInclusive && m.Score < maxScore) {
			if (minInclusive && m.Score >= minScore) || (!minInclusive && m.Score > minScore) {
				result = append(result, m.Member)
			}
		}
	}

	// 应用offset和count
	if offset > 0 {
		if offset >= int64(len(result)) {
			return []string{}, nil
		}
		result = result[offset:]
	}
	if count > 0 && count < int64(len(result)) {
		result = result[:count]
	}

	return result, nil
}

// ZRemRangeByRank 按排名范围删除成员
func (p *CacheMem) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	sorted := zset.getSortedMembers(true)
	toRemove := p.sliceRange(sorted, start, stop)

	var count int64
	for _, member := range toRemove {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZRemRangeByScore 按分数范围删除成员
func (p *CacheMem) ZRemRangeByScore(key, min, max string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	toRemove := make([]string, 0)
	for member, score := range zset.members {
		if (minInclusive && score >= minScore) || (!minInclusive && score > minScore) {
			if (maxInclusive && score <= maxScore) || (!maxInclusive && score < maxScore) {
				toRemove = append(toRemove, member)
			}
		}
	}

	var count int64
	for _, member := range toRemove {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZUnionStore 并集存储
func (p *CacheMem) ZUnionStore(destination string, keys ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}

	// 创建新的ZSet存储结果
	result := newZSet()

	for _, key := range keys {
		zset, exists := p.zsets[key]
		if !exists {
			continue
		}

		zset.mu.RLock()
		for member, score := range zset.members {
			existingScore, exists := result.members[member]
			if !exists {
				result.members[member] = score
			} else {
				// 并集：分数相加
				result.members[member] = existingScore + score
			}
		}
		zset.mu.RUnlock()
	}

	p.zsets[destination] = result
	return int64(len(result.members)), nil
}

// ZInterStore 交集存储
func (p *CacheMem) ZInterStore(destination string, keys ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}

	if len(keys) == 0 {
		return 0, nil
	}

	// 创建新的ZSet存储结果
	result := newZSet()

	// 获取第一个集合作为基准
	firstZset, exists := p.zsets[keys[0]]
	if !exists {
		p.zsets[destination] = result
		return 0, nil
	}

	firstZset.mu.RLock()
	defer firstZset.mu.RUnlock()

	// 遍历第一个集合的每个成员
	for member, score := range firstZset.members {
		sumScore := score
		inAll := true

		// 检查是否在其他所有集合中
		for i := 1; i < len(keys); i++ {
			zset, exists := p.zsets[keys[i]]
			if !exists {
				inAll = false
				break
			}

			zset.mu.RLock()
			otherScore, exists := zset.members[member]
			zset.mu.RUnlock()

			if !exists {
				inAll = false
				break
			}

			sumScore += otherScore
		}

		if inAll {
			result.members[member] = sumScore
		}
	}

	p.zsets[destination] = result
	return int64(len(result.members)), nil
}

// 辅助方法

// sliceRange 从sorted列表中提取start到stop范围的成员
func (p *CacheMem) sliceRange(sorted []ZSetMember, start, stop int64) []string {
	length := int64(len(sorted))
	if length == 0 {
		return []string{}
	}

	// 处理负索引
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}
	}

	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, sorted[i].Member)
	}

	return result
}

// sliceRangeWithScore 从sorted列表中提取start到stop范围的成员（包含分数）
func (p *CacheMem) sliceRangeWithScore(sorted []ZSetMember, start, stop int64) []ZSetMember {
	length := int64(len(sorted))
	if length == 0 {
		return []ZSetMember{}
	}

	// 处理负索引
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []ZSetMember{}
	}

	result := make([]ZSetMember, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, sorted[i])
	}

	return result
}

// parseScore 解析分数字符串（支持-inf, +inf, (score等Redis语法）
func (p *CacheMem) parseScore(scoreStr string) (float64, bool) {
	if scoreStr == "-inf" {
		return -1e308, true // 使用极小值表示负无穷
	}
	if scoreStr == "+inf" {
		return 1e308, true // 使用极大值表示正无穷
	}

	inclusive := true
	if len(scoreStr) > 0 && scoreStr[0] == '(' {
		inclusive = false
		scoreStr = scoreStr[1:]
	}

	var score float64
	_, err := fmt.Sscanf(scoreStr, "%f", &score)
	if err != nil {
		return 0, true
	}

	return score, inclusive
}
