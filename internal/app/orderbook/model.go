package orderbook

type Orderbook struct {
	Asks      any   `json:"asks" mapstructure:"asks"`
	Bids      any   `json:"bids" mapstructure:"bids"`
	Timestamp int64 `json:"timestamp" mapstructure:"timestamp"`
}
