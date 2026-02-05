package cache

import (
	"encoding/json"
	"math/rand"
	"time"
)

func (p *CacheMem) SAdd(key string, members ...string) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		p.data[key] = &Item{Data: "[]"}
		item = p.data[key]
	}

	var setMembers []string
	if err := json.Unmarshal([]byte(item.Data), &setMembers); err != nil {
		setMembers = make([]string, 0)
	}

	setMap := make(map[string]bool)
	for _, member := range setMembers {
		setMap[member] = true
	}

	addedCount := int64(0)
	for _, member := range members {
		if !setMap[member] {
			setMap[member] = true
			addedCount++
		}
	}

	newMembers := make([]string, 0, len(setMap))
	for member := range setMap {
		newMembers = append(newMembers, member)
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return addedCount, nil
}

func (p *CacheMem) SMembers(key string) ([]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return []string{}, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return []string{}, nil
	}

	return members, nil
}

func (p *CacheMem) SRem(key string, members ...string) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return 0, nil
	}

	var setMembers []string
	if err := json.Unmarshal([]byte(item.Data), &setMembers); err != nil {
		return 0, nil
	}

	removeMap := make(map[string]bool)
	for _, member := range members {
		removeMap[member] = true
	}

	newMembers := make([]string, 0)
	removedCount := int64(0)
	for _, member := range setMembers {
		if removeMap[member] {
			removedCount++
		} else {
			newMembers = append(newMembers, member)
		}
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return removedCount, nil
}

func (p *CacheMem) SRandMember(key string, count ...int64) ([]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return []string{}, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return []string{}, nil
	}

	if len(members) == 0 {
		return []string{}, nil
	}

	n := int64(1)
	if len(count) > 0 && count[0] > 0 {
		n = count[0]
	}

	if n >= int64(len(members)) {
		return members, nil
	}

	result := make([]string, 0, n)
	selected := make(map[int]bool)
	for int64(len(result)) < n {
		idx := rand.Intn(len(members))
		if !selected[idx] {
			selected[idx] = true
			result = append(result, members[idx])
		}
	}

	return result, nil
}

func (p *CacheMem) SPop(key string) (string, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return "", nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil || len(members) == 0 {
		return "", nil
	}

	idx := rand.Intn(len(members))
	popped := members[idx]
	newMembers := make([]string, 0, len(members)-1)
	for i, member := range members {
		if i != idx {
			newMembers = append(newMembers, member)
		}
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return popped, nil
}

func (p *CacheMem) SisMember(key, field string) (bool, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return false, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return false, nil
	}

	for _, member := range members {
		if member == field {
			return true, nil
		}
	}

	return false, nil
}
