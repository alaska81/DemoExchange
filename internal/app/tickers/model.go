package tickers

type Ticker struct {
	MarketID             string  `json:"market_id" mapstructure:"market_id"`
	Symbol               string  `json:"symbol" mapstructure:"symbol"`
	Timestamp            int64   `json:"timestamp" mapstructure:"timestamp"`
	DateTime             string  `json:"datetime" mapstructure:"datetime"`
	Bid                  float64 `json:"bid" mapstructure:"bid"`
	BidVolume            float64 `json:"bidVolume" mapstructure:"bidVolume"`
	Ask                  float64 `json:"ask" mapstructure:"ask"`
	AskVolume            float64 `json:"askVolume" mapstructure:"askVolume"`
	Open                 float64 `json:"open" mapstructure:"open"`
	High                 float64 `json:"high" mapstructure:"high"`
	Low                  float64 `json:"low" mapstructure:"low"`
	Close                float64 `json:"close" mapstructure:"close"`
	Last                 float64 `json:"last" mapstructure:"last"`
	BaseVolume           float64 `json:"baseVolume" mapstructure:"baseVolume"`
	QuoteVolume          float64 `json:"quoteVolume" mapstructure:"quoteVolume"`
	Change               float64 `json:"change" mapstructure:"change"`
	Percentage           float64 `json:"percentage" mapstructure:"percentage"`
	PriceMark            float64 `json:"price_mark" mapstructure:"price_mark"`
	PriceIndex           float64 `json:"price_index" mapstructure:"price_index"`
	LastFundingRate      float64 `json:"last_funding" mapstructure:"last_funding"`
	NextFundingTimestamp int64   `json:"next_funding_timestamp" mapstructure:"next_funding_timestamp"`
	NextFundingDateTime  string  `json:"next_funding_datetime" mapstructure:"next_funding_datetime"`
}

type Tickers map[string]Ticker
