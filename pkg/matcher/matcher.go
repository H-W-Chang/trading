package matcher

type Matcher interface {
	Match(o Order) (string, error)
}

const (
	Deal    string = "deal"
	Pending string = "add to pending list"
)

type QueryCondition struct {
	OrderID string
	Op      int8
	Price   float64
}

type OrderRepository interface {
	Query(condition QueryCondition) []Order
	Update(condition QueryCondition, orders []Order) error
	Insert(condition QueryCondition, order Order) error
	Delete(condition QueryCondition) error
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
	return nil
}
