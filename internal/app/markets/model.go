package markets

type Market struct {
	ID        string    `json:"id" mapstructure:"id"`
	Symbol    string    `json:"symbol" mapstructure:"symbol"`
	AltName   string    `json:"alt_name" mapstructure:"alt_name"`
	Base      string    `json:"base" mapstructure:"base"`
	Quote     string    `json:"quote" mapstructure:"quote"`
	Taker     float64   `json:"taker" mapstructure:"taker"`
	Maker     float64   `json:"maker" mapstructure:"maker"`
	Active    bool      `json:"active" mapstructure:"active"`
	Precision precision `json:"precision" mapstructure:"precision"`
	Limits    limit     `json:"limits" mapstructure:"limits"`
	IsSpot    bool      `json:"is_spot" mapstructure:"is_spot"`
	IsMargin  bool      `json:"is_margin" mapstructure:"is_margin"`
}

type precision struct {
	Amount int64 `json:"amount" mapstructure:"amount"`
	Price  int64 `json:"price" mapstructure:"price"`
	Cost   int64 `json:"cost" mapstructure:"cost"`
}

type limit struct {
	Amount band `json:"amount" mapstructure:"amount"`
	Price  band `json:"price" mapstructure:"price"`
	Cost   band `json:"cost" mapstructure:"cost"`
}

type band struct {
	Min float64 `json:"min" mapstructure:"min"`
	Max float64 `json:"max" mapstructure:"max"`
}

type Markets map[string]Market
