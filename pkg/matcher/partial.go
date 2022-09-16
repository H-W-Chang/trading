package matcher

type PartialMatcher struct {
	Ol OrderList
}

func (p *PartialMatcher) Match(newOrder Order) {
	switch newOrder.Op {
	case 0: //buy
		sol := p.Ol.FindByPrice(newOrder.Price)
		tempOrder := newOrder
		for _, so := range sol {
			if tempOrder.Volume <= 0 {
				break
			}
			if tempOrder.Volume >= so.Volume {
				tempOrder.Volume -= so.Volume
				err := p.Ol.DeleteOrder(so.OrderID)
				if err != nil {
					return
				}
			} else {
				so.Volume -= tempOrder.Volume
				tempOrder.Volume = 0
				err := p.Ol.UpdateOrder(so)
				if err != nil {
					return
				}
			}
		}
		if tempOrder.Volume > 0 {
			err := p.Ol.AddOrder(tempOrder)
			if err != nil {
				return
			}
		}
	case 1: //sell
	}

}
