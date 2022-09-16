package database

import "trading/pkg/order"

type DB interface {
	FindSellOrderByPrice(float64) []order.Order
}
type MatcherFactory struct {
}

func GetDB() DB {
	return &InMemoryDB{}
}
