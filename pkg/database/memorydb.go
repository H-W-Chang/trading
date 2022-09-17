package database

import (
	"sync"
	"trading/pkg/matcher"
)

type OrderQueue struct {
	queue         []matcher.Order
	totalVolume   int
	mux           sync.RWMutex
	inTransaction bool
}

type MemoryRepository map[int8]map[float64]*OrderQueue

var _ matcher.OrderRepository = &MemoryRepository{}

func (m MemoryRepository) Query(condition matcher.QueryCondition) []matcher.Order {
	if _, ok := m[condition.Op][condition.Price]; ok {
		return m[condition.Op][condition.Price].queue
	}
	return nil
}
func (m MemoryRepository) Update(condition matcher.QueryCondition, orders []matcher.Order) error {
	if m[condition.Op][condition.Price] == nil {
		m[condition.Op][condition.Price] = &OrderQueue{}
	}
	m[condition.Op][condition.Price].queue = orders
	return nil
}
func (m MemoryRepository) Insert(condition matcher.QueryCondition, order matcher.Order) error {
	if m[condition.Op][condition.Price] == nil {
		m[condition.Op][condition.Price] = &OrderQueue{}
	}
	m[condition.Op][condition.Price].queue = append(m[condition.Op][condition.Price].queue, order)
	return nil
}

func (m MemoryRepository) Delete(condition matcher.QueryCondition) error {
	if _, ok := m[condition.Op][condition.Price]; ok {
		m[condition.Op][condition.Price].queue = []matcher.Order{}
	}
	return nil
}
