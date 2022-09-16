package database

import (
	"trading/pkg/matcher"
)

type MemoryDB struct {
}

var _ matcher.OrderList = &MemoryDB{}

func (m *MemoryDB) FindByPrice(price float64) []matcher.Order {
	return nil

}
func (m *MemoryDB) DeleteOrder(string) error
func (m *MemoryDB) UpdateOrder(matcher.Order) error
func (m *MemoryDB) AddOrder(matcher.Order) error
