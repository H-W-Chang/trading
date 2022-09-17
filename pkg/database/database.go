package database

import "trading/pkg/matcher"

func NewRepository(dbType string) matcher.OrderRepository {
	switch dbType {
	case "memory":
		repo := make(MemoryRepository)
		repo[0] = make(map[float64]*OrderQueue)
		repo[1] = make(map[float64]*OrderQueue)
		return &repo
	}
	return nil
}
