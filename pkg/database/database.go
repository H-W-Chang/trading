package database

import (
	"trading/pkg/entity"
)

func NewRepository(dbType string) entity.PendingOrderRepository {
	switch dbType {
	case "memory":
		repo := &MemoryRepository{}
		repo.queue = make(map[int8]map[float64][]entity.Order)
		repo.queue[entity.Buy] = make(map[float64][]entity.Order)
		repo.queue[entity.Sell] = make(map[float64][]entity.Order)
		return repo
	}
	return nil
}
