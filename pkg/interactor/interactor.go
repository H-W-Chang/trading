package interactor

type Order struct {
	OrderID   string  `json:"orderID"`
	UserID    string  `json:"userID"`
	Item      string  `json:"item"`
	Op        int8    `json:"op"` //0 buy, 1 sell
	Volume    int     `json:"volume"`
	Price     float64 `json:"price"`
	MatchRule string  `json:"matchRule"`
}
type MatchRule interface {
	Match(order Order)
}
type Matcher struct {
	MatchRule string
}

func (m *Matcher) Match(order Order) {
	//TODO
}

func Match(order Order) {
	matcher := Matcher{MatchRule: order.MatchRule}
	matcher.Match(order)
	// var matchRule MatchRule
	// switch order.MatcherRule {
	// case "partial":
	// 	matchRule = &PartialMatchRule{}
	// 	matchRule.Match(order)
	// case "full":
	// 	matchRule.Match(order)
	// }
}
