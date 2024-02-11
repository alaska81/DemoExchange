package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/usecase/trade"
)

const (
	Timeout     = 5 * time.Second
	Limit   int = 100
)

//lint:ignore ST1005 strings capitalized
var ErrOrderAlreadyCancelled = errors.New("Order already cancelled")

func (uc *Usecase) NewOrder(ctx context.Context, accountUID entities.AccountUID, order entities.Order) (*entities.Order, error) {
	var err error

	order.NewUID()
	order.AccountUID = accountUID
	order.Status = entities.OrderStatusNew
	order.UpdateTS = order.CreateTS

	if err := uc.validateOrder(ctx, &order); err != nil {
		return nil, err
	}

	trader, err := uc.trade.Create(&order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("NewOrder:Create [%+v] error: %v", order, err))
		return nil, err
	}

	err = uc.openOrder(ctx, &order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("NewOrder:openOrder [%+v] error: %v", order, err))
		return nil, err
	}

	err = uc.order.InsertOrder(ctx, &order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("NewOrder:InsertOrder [%+v] error: %v", order, err))
		uc.closeOrder(ctx, &order)
		return nil, err
	}

	uc.trade.Set(&order)

	go func() {
		uc.chTraders <- trader
	}()

	return &order, nil
}

func (uc *Usecase) Process(ctx context.Context) {
	for {
		in := <-uc.chTraders

		trader, ok := in.(trade.Trader)
		if !ok {
			continue
		}

		go func() {
			ch := trader.Process(ctx)

			for {
				select {
				case <-ctx.Done():
					return
				case order, ok := <-ch:
					if !ok {
						return
					}

					if order.Status == entities.OrderStatusSuccess {
						if err := uc.executeOrder(ctx, order); err != nil {
							uc.log.Error(fmt.Sprintf("Process:executeOrder [%+v] error: %v", order, err))
							order.Status = entities.OrderStatusFailed
							order.Error = err.Error()
						}
					}

					if order.Status == entities.OrderStatusCancelled || order.Status == entities.OrderStatusFailed {
						if err := uc.closeOrder(ctx, order); err != nil {
							uc.log.Error(fmt.Sprintf("Process:closeOrder [%+v] error: %v", order, err))
							order.Status = entities.OrderStatusFailed
							order.Error = err.Error()
						}
					}

					uc.updateOrder(ctx, order)

					if order.Status == entities.OrderStatusSuccess || order.Status == entities.OrderStatusCancelled || order.Status == entities.OrderStatusFailed {
						uc.trade.Delete(order.OrderUID)
						return
					}
				}
			}
		}()
	}
}

func (uc *Usecase) GetOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error) {
	order, err := uc.order.SelectOrder(ctx, exchange, accountUID, orderUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("GetOrder:SelectOrder [account_uid: %v] [order_uid: %v] error: %v", accountUID, orderUID, err))
		return nil, err
	}

	return order, nil
}

func (uc *Usecase) CancelOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error) {
	order, err := uc.trade.Get(orderUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("CancelOrder:Get [account_uid: %v] [order_uid: %v] error: %v", accountUID, orderUID, err))
		return nil, err
	}

	if order.Status == entities.OrderStatusCancelled {
		return order, ErrOrderAlreadyCancelled
	}

	order.Status = entities.OrderStatusCancelled
	order.UpdateTS = entities.TS()

	return order, nil
}

func (uc *Usecase) OrdersList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, statuses []entities.OrderStatus, limit int) ([]*entities.Order, error) {
	if limit == 0 {
		limit = Limit
	}
	return uc.order.SelectOrders(ctx, exchange, accountUID, statuses, limit)
}

func (uc *Usecase) ProcessPendingOrders(ctx context.Context) error {
	orders, err := uc.order.SelectPendingOrders(ctx)
	if err != nil {
		uc.log.Error(fmt.Sprintf("ProcessPendingOrders:SelectPendingOrders error: %v", err))
		return err
	}

	uc.log.Info("ProcessPendingOrders: ", len(orders))

	for _, order := range orders {
		trader, err := uc.trade.Create(order)
		if err != nil {
			return err
		}

		uc.trade.Set(order)

		go func() {
			uc.chTraders <- trader
		}()

	}

	return nil
}

func (uc *Usecase) openOrder(ctx context.Context, order *entities.Order) error {
	return uc.holdBalance(ctx, order)
}

func (uc *Usecase) closeOrder(ctx context.Context, order *entities.Order) error {
	return uc.unholdBalance(ctx, order)
}

func (uc *Usecase) executeOrder(ctx context.Context, order *entities.Order) error {
	return uc.appendBalance(ctx, order)
}

func (uc *Usecase) validateOrder(ctx context.Context, order *entities.Order) error {
	account, err := uc.getAccountByUID(ctx, nil, order.AccountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("validateOrder:GetAccountByUID [AccountUID: %s] error: %v", order.AccountUID, err))
		return err
	}

	if account.PositionMode == entities.PositionModeOneway {
		order.PositionSide = entities.PositionSideBoth
	} else {
		if order.PositionSide != entities.PositionSideLong && order.PositionSide != entities.PositionSideShort {
			return apperror.ErrInvalidPositionSide
		}
	}

	return nil
}

func (uc *Usecase) updateOrder(ctx context.Context, order *entities.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			order.UpdateTS = entities.TS()
			if err := uc.order.UpdateOrder(ctx, order); err != nil {
				uc.log.Error(fmt.Sprintf("updateOrder:UpdateOrder [%+v] error: %v", order, err))
				time.Sleep(Timeout)
				continue
			}
			return
		}
	}
}

func (uc *Usecase) checkPresentPendingOrders(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol *entities.Symbol) error {
	orders, err := uc.order.SelectPendingOrdersBySymbol(ctx, exchange, accountUID, symbol)
	if err != nil {
		uc.log.Error(fmt.Sprintf("checkPresentPendingOrders:SelectPendingOrdersBySymbol [account_uid: %v, symbol: %v] error: %v", accountUID, symbol, err))
		return apperror.ErrRequestError
	}

	if len(orders) > 0 {
		return apperror.ErrOpenOrdersExists
	}

	return nil
}
