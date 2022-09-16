package database

import "trading/pkg/order"

type InMemoryDB struct {
}

func (m *InMemoryDB) FindSellOrderByPrice(float64) []order.Order {

}
