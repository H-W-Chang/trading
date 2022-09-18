package matcher

import (
	"trading/pkg/entity"
)

type Matcher interface {
	Match(o entity.Order) (string, error)
}

const (
	Deal    string = "deal"
	Pending string = "add to pending list"
)
