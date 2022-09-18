package database

import (
	"sync"
	"trading/pkg/entity"
)

type OrderQueue struct {
	queue         []entity.Order
	totalVolume   int
	mux           sync.RWMutex
	inTransaction bool
}

type MemoryRepository map[int8]map[float64]*OrderQueue

var _ entity.PendingOrderRepository = &MemoryRepository{}

func (m MemoryRepository) Query(condition entity.QueryCondition) []entity.Order {
	if _, ok := m[condition.Op][condition.Price]; ok {
		return m[condition.Op][condition.Price].queue
	}
	return nil
}
func (m MemoryRepository) Update(condition entity.QueryCondition, orders []entity.Order) error {
	if m[condition.Op][condition.Price] == nil {
		m[condition.Op][condition.Price] = &OrderQueue{}
	}
	m[condition.Op][condition.Price].queue = orders
	return nil
}
func (m MemoryRepository) Insert(condition entity.QueryCondition, order entity.Order) error {
	if m[condition.Op][condition.Price] == nil {
		m[condition.Op][condition.Price] = &OrderQueue{}
	}
	m[condition.Op][condition.Price].queue = append(m[condition.Op][condition.Price].queue, order)
	return nil
}

func (m MemoryRepository) Delete(condition entity.QueryCondition) error {
	switch condition.Op {
	case entity.Buy:
		if m[entity.Buy][condition.Price] != nil {
			delete(m[entity.Buy], condition.Price)
		}
	case entity.Sell:
		if m[entity.Sell][condition.Price] != nil {
			delete(m[entity.Sell], condition.Price)
		}
	case entity.All:
		if m[entity.Buy][condition.Price] != nil {
			delete(m[entity.Buy], condition.Price)
		}
		if m[entity.Sell][condition.Price] != nil {
			delete(m[entity.Sell], condition.Price)
		}
	}
	if _, ok := m[condition.Op][condition.Price]; ok {
		m[condition.Op][condition.Price].queue = []entity.Order{}
	}
	return nil
}
