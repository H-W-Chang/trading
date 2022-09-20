package database

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"trading/pkg/entity"
)

type MemoryRepository struct {
	queue         map[int8]map[float64][]entity.Order
	queueMutex    sync.RWMutex
	metaMutex     sync.RWMutex
	lockId        string
	inTransaction bool
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
	m.metaMutex.RLock()
	// log.Printf("query routine id: %v, trans: %v, order lockId: %v, condition lockId: %v", goid(), m.inTransaction, m.lockId, condition.LockId)
	if m.inTransaction && m.lockId != "" && m.lockId == condition.LockId {
		m.metaMutex.RUnlock()
	} else {
		m.metaMutex.RUnlock()
		m.queueMutex.RLock()
		defer m.queueMutex.RUnlock()
	}
	if m.queue == nil {
		return nil
	}
	if queue, ok := m.queue[condition.Op][condition.Price]; ok {
		return queue
	}

	return nil
}
func (m *MemoryRepository) Update(condition entity.QueryCondition, orders []entity.Order) error {
	m.metaMutex.RLock()
	// log.Printf("update routine id: %v, trans: %v, order lockId: %v, condition lockId: %v", goid(), m.inTransaction, m.lockId, condition.LockId)
	if m.inTransaction && m.lockId != "" && m.lockId == condition.LockId {
		m.metaMutex.RUnlock()
	} else {
		m.metaMutex.RUnlock()
		m.queueMutex.Lock()
		defer m.queueMutex.Unlock()
	}

	m.queue[condition.Op][condition.Price] = orders
	// log.Printf("after update: %+v", m.queue[condition.Op][condition.Price])

	return nil
}
func (m *MemoryRepository) Insert(condition entity.QueryCondition, order entity.Order) error {
	m.metaMutex.RLock()
	// log.Printf("insert routine id: %v, trans: %v, order lockId: %v, condition lockId: %v", goid(), m.inTransaction, m.lockId, condition.LockId)
	if m.inTransaction && m.lockId != "" && m.lockId == condition.LockId {
		// log.Printf("routine id: %v, insert in transaction", goid())
		m.metaMutex.RUnlock()
	} else {
		// log.Printf("routine id: %v, insert lock", goid())
		m.metaMutex.RUnlock()
		m.queueMutex.Lock()
		defer m.queueMutex.Unlock()
	}
	if _, ok := m.queue[condition.Op][condition.Price]; !ok {
		m.queue[condition.Op][condition.Price] = []entity.Order{}
	}
	m.queue[condition.Op][condition.Price] = append(m.queue[condition.Op][condition.Price], order)
	// log.Printf("after insert: %+v", m.queue[condition.Op][condition.Price])
	return nil

}

func (m *MemoryRepository) Delete(condition entity.QueryCondition) error {
	if condition.OrderID == "" {
		if condition.Op == entity.All {
			m.metaMutex.RLock()
			if m.inTransaction && m.lockId != "" && m.lockId == condition.LockId {
				m.metaMutex.RUnlock()
			} else {
				m.metaMutex.RUnlock()
				m.queueMutex.Lock()
				defer m.queueMutex.Unlock()
			}
			for i := range m.queue {
				m.queue[i][condition.Price] = []entity.Order{}
			}
		} else {
			m.metaMutex.RLock()
			if m.inTransaction && m.lockId != "" && m.lockId == condition.LockId {
				m.metaMutex.RUnlock()
			} else {
				m.metaMutex.RUnlock()
				m.queueMutex.Lock()
				defer m.queueMutex.Unlock()
			}
			m.queue[condition.Op][condition.Price] = []entity.Order{}
		}
	}
	return nil
}

func (m *MemoryRepository) Lock(condition entity.QueryCondition) string {
	h := sha1.New()
	io.WriteString(h, condition.OrderID)
	m.queueMutex.Lock()
	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()
	m.inTransaction = true
	m.lockId = base64.URLEncoding.EncodeToString(h.Sum(nil))
	// log.Printf("lock routine id: %v, trans: %v, order lockId: %v", goid(), m.inTransaction, m.lockId)
	return m.lockId
}

func (m *MemoryRepository) Unlock(condition entity.QueryCondition) error {
	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()
	if m.inTransaction && m.lockId != "" {
		// log.Printf("unlock routine id: %v, trans: %v, order lockId: %v", goid(), m.inTransaction, m.lockId)
		m.inTransaction = false
		m.lockId = ""
		m.queueMutex.Unlock()
		return nil
	}

	return errors.New("unlock failed")
}
