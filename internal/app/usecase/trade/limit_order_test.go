package trade

import (
	"DemoExchange/internal/app/entities"
	"testing"
)

func TestLimitOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		order   *LimitOrder
		wantErr error
	}{
		{
			name: "Valid order",
			order: &LimitOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USD",
				Side:     entities.OrderSideBuy,
				Amount:   1.5,
				Price:    60000,
			},
			wantErr: nil,
		},
		{
			name: "Invalid exchange",
			order: &LimitOrder{
				Exchange: "invalidExchange",
				Symbol:   "BTC/USD",
				Side:     entities.OrderSideBuy,
				Amount:   1.5,
				Price:    60000,
			},
			wantErr: ErrExchangeIsNotValid,
		},
		{
			name: "Empty symbol",
			order: &LimitOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "",
				Side:     entities.OrderSideBuy,
				Amount:   1.5,
				Price:    60000,
			},
			wantErr: ErrSymbolIsNotValid,
		},
		{
			name: "Invalid side",
			order: &LimitOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     "invalid",
				Amount:   1.5,
				Price:    60000,
			},
			wantErr: ErrOrderSideIsNotValid,
		},
		{
			name: "ZeroAmount",
			order: &LimitOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     entities.OrderSideBuy,
				Amount:   0,
				Price:    60000,
			},
			wantErr: ErrAmountIsNotValid,
		},
		{
			name: "ZeroPrice",
			order: &LimitOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     entities.OrderSideBuy,
				Amount:   1.5,
				Price:    0,
			},
			wantErr: ErrPriceIsNotValid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.order.Validate(); err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
