package matcher

type PartialMatcher struct {
	Or OrderRepository
}

func (p *PartialMatcher) Match(newOrder Order) (string, error) {
	price := newOrder.Price
	condition := QueryCondition{
		Op:    newOrder.Op,
		Price: price,
	}
	orderQueue := p.Or.Query(condition)
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
	p.Or.Update(condition, orderQueueCopy)
	if newOrderCopy.Volume > 0 {
		p.Or.Insert(condition, newOrderCopy)
		// orders.CommitTx(price)
		return Pending, nil
	}
	// orders.CommitTx(price)
	return Deal, nil
}
