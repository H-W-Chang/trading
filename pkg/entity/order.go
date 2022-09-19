package entity

import "errors"

type QueryCondition struct {
	OrderID string
	Op      int8
	Price   float64
	LockId  string
}

const (
	Buy  int8 = 0
	Sell int8 = 1
	All  int8 = 99
)

type PendingOrderRepository interface {
	Query(condition QueryCondition) []Order
	Update(condition QueryCondition, orders []Order) error
	Insert(condition QueryCondition, order Order) error
	Delete(condition QueryCondition) error
	Lock(condition QueryCondition) string
	Unlock(condition QueryCondition) error
}
type Order struct {
	OrderID   string  `json:"orderID"`
	UserID    string  `json:"userID"`
	Item      string  `json:"item"`
	Op        int8    `json:"op"` //0 buy, 1 sell
	Volume    int     `json:"volume"`
	Price     float64 `json:"price"`
	MatchRule string  `json:"matchRule"`
}

func (o *Order) Validate() error {
	if o.Op < Buy || o.Op > Sell {
		return errors.New("op must between 0 and 1")
	}
	if o.Price < 0 {
		return errors.New("price must >= 0")
	}
	if o.Volume <= 0 {
		return errors.New("volume must > 0")
	}
	return nil
}

func GetOppositeOp(op int8) int8 {
	switch op {
	case Buy:
		return Sell
	case Sell:
		return Buy
	default:
		return -1
	}
}
