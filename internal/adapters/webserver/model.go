package webserver

type Responce struct {
	Success bool        `json:"success"`
	Return  interface{} `json:"return,omitempty"`
	Error   string      `json:"error,omitempty"`
	Time    string      `json:"time"`
}

type CreateTokenRequest struct {
	// Exchange string `json:"exchange"`
	Service string `json:"service"`
	UserID  string `json:"user_id"`
}

type DisableTokenRequest struct {
	// Exchange string `json:"exchange"`
	Token string `json:"token"`
}

type DepositRequest struct {
	Exchange string  `json:"exchange"`
	Coin     string  `json:"coin"`
	Amount   float64 `json:"amount"`
}

type WithdrawRequest struct {
	Exchange string  `json:"exchange"`
	Coin     string  `json:"coin"`
	Amount   float64 `json:"amount"`
}

type OrderCreateRequest struct {
	Exchange     string  `json:"exchange"`
	Symbol       string  `json:"symbol"`
	Type         string  `json:"type"`
	PositionSide string  `json:"position_side"`
	Side         string  `json:"side"`
	Amount       float64 `json:"amount"`
	Price        float64 `json:"price"`
}

type OrderRequest struct {
	Exchange string `json:"exchange"`
	OrderUID string `json:"order_uid"`
}

type PositionModeRequest struct {
	Exchange string `json:"exchange"`
	Mode     string `json:"mode"`
}

type PositionTypeRequest struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
	Type     string `json:"type"`
}

type PositionLeverageRequest struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
	Leverage int32  `json:"leverage"`
}
