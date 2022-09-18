package matcher

import (
	"trading/pkg/entity"
)

type PartialMatcher struct {
	Or entity.PendingOrderRepository
}

func (p *PartialMatcher) Match(newOrder entity.Order) (string, error) {
	price := newOrder.Price
	condition := entity.QueryCondition{
		Op:    newOrder.Op,
		Price: price,
	}
	orderQueue := p.Or.Query(entity.QueryCondition{Op: entity.GetOppositeOp(newOrder.Op), Price: price})
	if orderQueue == nil {
		p.Or.Insert(condition, newOrder)
		// orders.CommitTx(price)
		return Pending, nil
	}
	orderQueueCopy := orderQueue
	newOrderCopy := newOrder
	for i := range orderQueue {
		if newOrderCopy.Volume <= 0 {
			break
		}
		if newOrderCopy.Volume >= orderQueue[i].Volume {
			newOrderCopy.Volume -= orderQueue[i].Volume
			orderQueueCopy = orderQueueCopy[1:]
		} else {
			orderQueueCopy[0].Volume -= newOrderCopy.Volume
			newOrderCopy.Volume = 0
		}
	}
	p.Or.Update(entity.QueryCondition{Op: entity.GetOppositeOp(newOrder.Op), Price: price}, orderQueueCopy)
	if newOrderCopy.Volume > 0 {
		p.Or.Insert(condition, newOrderCopy)
		// orders.CommitTx(price)
		return Pending, nil
	}
	// orders.CommitTx(price)
	return Deal, nil
}
