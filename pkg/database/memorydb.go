package database

import (
	"crypto/sha1"
	"errors"
	"io"
	"sync"
	"trading/pkg/entity"
)

type OrderQueue struct {
	queue         []entity.Order
	mux           sync.RWMutex
	lockId        string
	inTransaction bool
}

type MemoryRepository map[int8]map[float64]*OrderQueue //op, price

var _ entity.PendingOrderRepository = &MemoryRepository{}

func (m MemoryRepository) Query(condition entity.QueryCondition) []entity.Order {
	if queue, ok := m[condition.Op][condition.Price]; ok {
		if queue == nil {
			return nil
		}
		if queue.inTransaction && queue.lockId == condition.LockId {
			return queue.queue
		} else {
			queue.mux.RLock()
			defer queue.mux.RUnlock()
			return queue.queue
		}
	}
	return nil
}
func (m MemoryRepository) Update(condition entity.QueryCondition, orders []entity.Order) error {
	if orderQueue, ok := m[condition.Op][condition.Price]; ok {
		if orderQueue == nil {
			orderQueue = &OrderQueue{}
			orderQueue.queue = orders
		} else {
			if orderQueue.inTransaction && orderQueue.lockId == condition.LockId {
				orderQueue.queue = orders
			} else {
				orderQueue.mux.Lock()
				defer orderQueue.mux.Unlock()
				orderQueue.queue = orders
			}
		}
	} else {
		m[condition.Op][condition.Price] = &OrderQueue{}
		m[condition.Op][condition.Price].queue = orders
	}
	return nil
}
func (m MemoryRepository) Insert(condition entity.QueryCondition, order entity.Order) error {
	if orderQueue, ok := m[condition.Op][condition.Price]; ok {
		if orderQueue == nil {
			orderQueue = &OrderQueue{}
			orderQueue.queue = append(orderQueue.queue, order)
		} else {
			if orderQueue.inTransaction && orderQueue.lockId == condition.LockId {
				orderQueue.queue = append(orderQueue.queue, order)
			} else {
				orderQueue.mux.Lock()
				defer orderQueue.mux.Unlock()
				orderQueue.queue = append(orderQueue.queue, order)
			}
		}
	} else {
		m[condition.Op][condition.Price] = &OrderQueue{}
		m[condition.Op][condition.Price].queue = append(m[condition.Op][condition.Price].queue, order)
	}
	return nil
}

func (m MemoryRepository) Delete(condition entity.QueryCondition) error {
	if condition.OrderID == "" {
		if condition.Op == entity.All {
			for _, opMap := range m {
				if orderQueue, ok := opMap[condition.Price]; ok {
					if orderQueue != nil {
						if orderQueue.inTransaction && orderQueue.lockId == condition.LockId {
							delete(m[entity.Buy], condition.Price)
						} else {
							orderQueue.mux.Lock()
							defer orderQueue.mux.Unlock()
							delete(m[entity.Buy], condition.Price)
						}
					}
				}
			}
		} else {
			if orderQueue, ok := m[condition.Op][condition.Price]; ok {
				if orderQueue != nil {
					if orderQueue.inTransaction && orderQueue.lockId == condition.LockId {
						delete(m[condition.Op], condition.Price)
					} else {
						orderQueue.mux.Lock()
						defer orderQueue.mux.Unlock()
						delete(m[condition.Op], condition.Price)
					}
				}
			}
		}
	}
	return nil
}

func (m MemoryRepository) Lock(condition entity.QueryCondition) string {
	var orderQueue *OrderQueue
	var ok bool
	if orderQueue, ok = m[condition.Op][condition.Price]; ok {
		if orderQueue == nil {
			orderQueue = &OrderQueue{}
		} else {
			orderQueue = m[condition.Op][condition.Price]
		}
	} else {
		m[condition.Op][condition.Price] = &OrderQueue{}
		orderQueue = m[condition.Op][condition.Price]
	}
	h := sha1.New()
	io.WriteString(h, condition.OrderID)
	orderQueue.mux.Lock()
	orderQueue.inTransaction = true
	orderQueue.lockId = string(h.Sum(nil))
	return orderQueue.lockId
}

func (m MemoryRepository) Unlock(condition entity.QueryCondition) error {
	if orderQueue, ok := m[condition.Op][condition.Price]; ok {
		if orderQueue != nil && orderQueue.inTransaction && orderQueue.lockId != "" {
			orderQueue.mux.Unlock()
			orderQueue.inTransaction = false
			orderQueue.lockId = ""
			return nil
		}
	}
	return errors.New("unlock failed")
}
