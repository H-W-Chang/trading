package matcher

import (
	"trading/pkg/database"
	"trading/pkg/order"
)

type PartialMatcher struct {
}

func (p *PartialMatcher) Match(newOrder order.Order) {
	db := database.GetDB()
	switch newOrder.Op {
	case 0: //buy
		err := db.BeginTx()
		sol := db.FindSellOrderByPrice(newOrder.Price)
		tempOrder := newOrder
		for _, so := range sol {
			if tempOrder.Volume <= 0 {
				break
			}
			if tempOrder.Volume >= so.Volume {
				tempOrder.Volume -= so.Volume
				err := db.DeleteSellOrder(so.OrderID)
				if err != nil {
					db.RollbackTx()
					return
				}
			} else {
				so.Volume -= tempOrder.Volume
				tempOrder.Volume = 0
				err := db.UpdateSellOrder(so)
				if err != nil {
					db.RollbackTx()
					return
				}
			}
		}
		if tempOrder.Volume > 0 {
			err := db.AddBuyOrder(tempOrder)
			if err != nil {
				db.RollbackTx()
				return
			}
		}
		db.CommitTx()
	case 1: //sell
	}

}
