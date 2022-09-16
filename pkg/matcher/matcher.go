package matcher

type Matcher interface {
	Match(o Order)
}

type OrderList interface {
	FindByPrice(float64) []Order
	DeleteOrder(string) error
	UpdateOrder(Order) error
	AddOrder(Order) error
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
