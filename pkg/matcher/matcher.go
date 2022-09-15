package matcher

import "trading/pkg/order"

type Matcher interface {
	Match(o order.Order)
}
