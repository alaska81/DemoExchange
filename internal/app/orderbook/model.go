package orderbook

type Orderbook struct {
	Asks      interface{} `json:"asks" mapstructure:"asks"`
	Bids      interface{} `json:"bids" mapstructure:"bids"`
	Timestamp int64       `json:"timestamp" mapstructure:"timestamp"`
}
