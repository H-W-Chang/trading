package database

import (
	"trading/pkg/entity"
)

func NewRepository(dbType string) entity.PendingOrderRepository {
	switch dbType {
	case "memory":
		repo := &MemoryRepository{}
		return repo
	}
	return nil
}
