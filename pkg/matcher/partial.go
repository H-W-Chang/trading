package matcher

import (
	"trading/pkg/entity"
)

type PartialMatcher struct {
	Or entity.PendingOrderRepository
}

func (p *PartialMatcher) Match(newOrder entity.Order) (string, error) {
	price := newOrder.Price
	condition := entity.QueryCondition{OrderID: newOrder.OrderID, Op: newOrder.Op, Price: price}
	oppositeCond := entity.QueryCondition{OrderID: newOrder.OrderID, Op: entity.GetOppositeOp(newOrder.Op), Price: price}
	lockId := p.Or.Lock(oppositeCond)
	defer p.Or.Unlock(oppositeCond)
	oppositeCond.LockId = lockId
	condition.LockId = lockId
	orderQueue := p.Or.Query(oppositeCond)
	if orderQueue == nil {
		p.Or.Insert(condition, newOrder)
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
	p.Or.Update(oppositeCond, orderQueueCopy)
	if newOrderCopy.Volume > 0 {
		p.Or.Insert(condition, newOrderCopy)
		return Pending, nil
	}
	return Deal, nil
}
