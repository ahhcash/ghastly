package storage

import (
	"math/rand"
)

const (
	maxLevel = 32
	p        = 0.25
)

type SkipNode struct {
	key   string
	value []byte
	next  []*SkipNode
}

type SkipList struct {
	head   *SkipNode
	level  int
	length int
}

func NewSkipList() *SkipList {
	return &SkipList{
		head: &SkipNode{
			next: make([]*SkipNode, maxLevel),
		},
		level: 0,
	}
}

func (s *SkipList) randomLevel() int {
	level := 0
	for rand.Float64() < p && level < maxLevel-1 {
		level++
	}
	return level
}

func (s *SkipList) Search(key string) ([]byte, bool) {
	current := s.head
	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && current.key < key {
			current = current.next[i]
		}
	}

	if current != nil && current.key == key {
		return current.value, true
	}

	return nil, false
}

func (s *SkipList) Iterator() func() (string, []byte, bool) {
	curr := s.head.next[0]
	return func() (string, []byte, bool) {
		if curr == nil {
			return "", nil, false
		}
		key := curr.key
		vector := curr.value
		curr = curr.next[0]
		return key, vector, true
	}
}

func (s *SkipList) Insert(key string, value []byte) {
	update := make([]*SkipNode, maxLevel)
	current := s.head

	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i]
		}
		update[i] = current
	}

	current = current.next[0]

	if current != nil && current.key == key {
		current.value = value
		return
	}

	newLevel := s.randomLevel()

	if newLevel > s.level {
		for i := s.level + 1; i <= newLevel; i++ {
			update[i] = s.head
		}
		s.level = newLevel
	}

	newNode := &SkipNode{
		key:   key,
		value: value,
		next:  make([]*SkipNode, newLevel+1),
	}

	for i := 0; i <= newLevel; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}

	s.length++

}
