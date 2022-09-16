package matcher

import (
	"trading/pkg/order"
)

type Matcher interface {
	Match(o order.Order)
}

func CreateMatcher(matchRule string) Matcher {
	switch matchRule {
	case "partial":
		return &PartialMatcher{}
	}
	return nil
}
