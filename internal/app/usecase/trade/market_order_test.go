package trade

import (
	"testing"

	"DemoExchange/internal/app/entities"

	"github.com/stretchr/testify/assert"
)

func TestMarketOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		order   *MarketOrder
		errType error
	}{
		{
			name: "ValidOrder",
			order: &MarketOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     entities.OrderSideBuy,
				Amount:   1,
			},
			errType: nil,
		},
		{
			name: "InvalidExchange",
			order: &MarketOrder{
				Exchange: "invalid",
				Symbol:   "BTC/USDT",
				Side:     entities.OrderSideBuy,
				Amount:   1,
			},
			errType: ErrExchangeIsNotValid,
		},
		{
			name: "EmptySymbol",
			order: &MarketOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "",
				Side:     entities.OrderSideBuy,
				Amount:   1,
			},
			errType: ErrSymbolIsNotValid,
		},
		{
			name: "InvalidSide",
			order: &MarketOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     "invalid",
				Amount:   1,
			},
			errType: ErrOrderSideIsNotValid,
		},
		{
			name: "ZeroAmount",
			order: &MarketOrder{
				Exchange: entities.ExchangeSpot,
				Symbol:   "BTC/USDT",
				Side:     entities.OrderSideBuy,
				Amount:   0,
			},
			errType: ErrAmountIsNotValid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			assert.ErrorIs(t, err, tt.errType)
		})
	}
}
