package database

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"trading/pkg/entity"
)

//	type OrderQueue struct {
//		queue         []entity.Order
//		mux           sync.RWMutex
//		lockId        string
//		inTransaction bool
//	}
type OrderQueue struct {
	queue         map[float64][]entity.Order
	mux           sync.RWMutex
	lockId        atomic.Value
	inTransaction atomic.Bool
}
type MemoryRepository struct {
	buyOrderQueue  OrderQueue
	sellOrderQueue OrderQueue
}

var _ entity.PendingOrderRepository = &MemoryRepository{}

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
func (m *MemoryRepository) Query(condition entity.QueryCondition) []entity.Order {
	var orderQueue *OrderQueue
	switch condition.Op {
	case entity.Buy:
		orderQueue = &m.buyOrderQueue
	case entity.Sell:
		orderQueue = &m.sellOrderQueue
	}

	log.Printf("routine id: %v, trans: %v, order lockId: %v, condition lockId: %v, *orderQueue: %p", goid(), orderQueue.inTransaction.Load(), orderQueue.lockId.Load().(string), condition.LockId, orderQueue)
	if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
	} else {
		orderQueue.mux.RLock()
		defer orderQueue.mux.RUnlock()

	}
	if orderQueue.queue == nil {
		return nil
	}
	if queue, ok := orderQueue.queue[condition.Price]; ok {
		return queue
	}
	return nil
}
func (m *MemoryRepository) Update(condition entity.QueryCondition, orders []entity.Order) error {
	var orderQueue *OrderQueue
	switch condition.Op {
	case entity.Buy:
		orderQueue = &m.buyOrderQueue
	case entity.Sell:
		orderQueue = &m.sellOrderQueue
	}
	if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
	} else {
		orderQueue.mux.Lock()
		defer orderQueue.mux.Unlock()
	}
	if orderQueue.queue == nil {
		orderQueue.queue = make(map[float64][]entity.Order)
	}
	orderQueue.queue[condition.Price] = orders

	return nil
}
func (m *MemoryRepository) Insert(condition entity.QueryCondition, order entity.Order) error {
	var orderQueue *OrderQueue
	switch condition.Op {
	case entity.Buy:
		orderQueue = &m.buyOrderQueue
	case entity.Sell:
		orderQueue = &m.sellOrderQueue
	}

	log.Printf("routine id: %v, trans: %v, order lockId: %v, condition lockId: %v, *orderQueue: %p", goid(), orderQueue.inTransaction.Load(), orderQueue.lockId.Load().(string), condition.LockId, orderQueue)
	if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
	} else {
		orderQueue.mux.Lock()
		defer orderQueue.mux.Unlock()
	}
	if orderQueue.queue == nil {
		orderQueue.queue = make(map[float64][]entity.Order)
	}
	if _, ok := orderQueue.queue[condition.Price]; !ok {
		orderQueue.queue[condition.Price] = []entity.Order{}
	}
	orderQueue.queue[condition.Price] = append(orderQueue.queue[condition.Price], order)
	return nil

}

func (m *MemoryRepository) Delete(condition entity.QueryCondition) error {
	if condition.OrderID == "" {
		if condition.Op == entity.All {
			//delete buy order queue
			var orderQueue *OrderQueue = &m.buyOrderQueue

			if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
			} else {
				orderQueue.mux.Lock()
				defer orderQueue.mux.Unlock()
			}
			if orderQueue.queue == nil {
				return nil
			}
			if _, ok := orderQueue.queue[condition.Price]; !ok {
				return nil
			}
			orderQueue.queue[condition.Price] = []entity.Order{}
			//delete sell order queue
			orderQueue = &m.sellOrderQueue

			if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
			} else {
				orderQueue.mux.Lock()
				defer orderQueue.mux.Unlock()
			}
			if orderQueue.queue == nil {
				return nil
			}
			if _, ok := orderQueue.queue[condition.Price]; !ok {
				return nil
			}
			orderQueue.queue[condition.Price] = []entity.Order{}
		} else {
			var orderQueue *OrderQueue
			switch condition.Op {
			case entity.Buy:
				orderQueue = &m.buyOrderQueue
			case entity.Sell:
				orderQueue = &m.sellOrderQueue
			}
			if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) == condition.LockId {
			} else {
				orderQueue.mux.Lock()
				defer orderQueue.mux.Unlock()
			}
			if orderQueue.queue == nil {
				return nil
			}
			if _, ok := orderQueue.queue[condition.Price]; !ok {
				return nil
			}
			orderQueue.queue[condition.Price] = []entity.Order{}
		}
	}
	return nil
}

func (m *MemoryRepository) Lock(condition entity.QueryCondition) string {
	var orderQueue *OrderQueue
	switch condition.Op {
	case entity.Buy:
		orderQueue = &m.buyOrderQueue
	case entity.Sell:
		orderQueue = &m.sellOrderQueue
	}
	h := sha1.New()
	io.WriteString(h, condition.OrderID)
	orderQueue.mux.Lock()
	orderQueue.inTransaction.Store(true)
	orderQueue.lockId.Store(base64.URLEncoding.EncodeToString(h.Sum(nil)))
	return orderQueue.lockId.Load().(string)
}

func (m *MemoryRepository) Unlock(condition entity.QueryCondition) error {
	var orderQueue *OrderQueue
	switch condition.Op {
	case entity.Buy:
		orderQueue = &m.buyOrderQueue
	case entity.Sell:
		orderQueue = &m.sellOrderQueue
	}
	if orderQueue.inTransaction.Load() && orderQueue.lockId.Load().(string) != "" {
		orderQueue.inTransaction.Store(false)
		orderQueue.lockId.Store("")
		orderQueue.mux.Unlock()
		return nil
	}

	return errors.New("unlock failed")
}
