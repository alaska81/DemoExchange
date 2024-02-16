package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/usecase/orders"
)

const (
	Timeout     = 5 * time.Second
	Limit   int = 100
)

var muOrder = &sync.RWMutex{}

func (uc *Usecase) SetOrder(ctx context.Context, o *entities.Order) error {
	var err error

	order, err := orders.NewOrder(ctx, uc, o)
	if err != nil {
		return err
	}
	if err := order.Validate(); err != nil {
		return err
	}

	err = uc.order.WithTx(ctx, func(ctx context.Context) error {
		muOrder.Lock()
		defer muOrder.Unlock()

		if err := order.HoldBalance(ctx, uc, uc.log); err != nil {
			uc.log.Error(fmt.Sprintf("SetOrder:HoldBalance [%+v] error: %v", *order, err))
			return err
		}

		if err := uc.saveOrder(ctx, order.GetOrder()); err != nil {
			uc.log.Error(fmt.Sprintf("SetOrder:saveOrder [%+v] error: %v", *order, err))
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	uc.cache.Set(order.GetOrder())

	go func() {
		uc.chOrders <- order
	}()

	return nil
}

func (uc *Usecase) Process(ctx context.Context) {
	for {
		in := <-uc.chOrders

		order, ok := in.(Order)
		if !ok {
			continue
		}

		go func() {
			defer func() {
				uc.cache.Delete(order.GetOrder().OrderUID)
			}()

			ch, err := order.Process(ctx)
			if err != nil {
				uc.log.Error(fmt.Sprintf("Process:Process [%+v] error: %v", order, err))
				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case status, ok := <-ch:
					uc.log.Info(fmt.Sprintf("Process:channel [%+v]: %v", status, ok))
					if !ok {
						return
					}

					muOrder.Lock()
					o := order.GetOrder()

					if status == entities.OrderStatusSuccess {
						if err := order.AppendBalance(ctx, uc, uc.log); err != nil {
							uc.log.Error(fmt.Sprintf("Process:AppendBalance [%+v] error: %v", *o, err))
							status = entities.OrderStatusFailed
							o.Error = err.Error()
						}
					}

					if status == entities.OrderStatusCancelled || status == entities.OrderStatusFailed {
						if err := order.UnholdBalance(ctx, uc, uc.log); err != nil {
							uc.log.Error(fmt.Sprintf("Process:HoldBalance [%+v] error: %v", *o, err))
							status = entities.OrderStatusFailed
							o.Error = err.Error()
						}
					}

					o.Status = status

					uc.updateOrder(ctx, o)
					muOrder.Unlock()

					if status == entities.OrderStatusSuccess || status == entities.OrderStatusCancelled || status == entities.OrderStatusFailed {
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
	order, ok := uc.cache.Get(orderUID)
	if !ok {
		err := apperror.ErrOrderNotFound
		uc.log.Error(fmt.Sprintf("CancelOrder:Get [account_uid: %v] [order_uid: %v] error: %v", accountUID, orderUID, err))
		return nil, err
	}

	if order.Status == entities.OrderStatusCancelled {
		return order, apperror.ErrOrderAlreadyCancelled
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
	pendingOrders, err := uc.order.SelectPendingOrders(ctx)
	if err != nil {
		uc.log.Error(fmt.Sprintf("ProcessPendingOrders:SelectPendingOrders error: %v", err))
		return err
	}

	uc.log.Info("ProcessPendingOrders: ", len(pendingOrders))

	for _, o := range pendingOrders {
		order, err := orders.NewOrder(ctx, uc, o)
		if err != nil {
			return err
		}

		uc.cache.Set(order.GetOrder())

		go func() {
			uc.chOrders <- order
		}()

	}

	return nil
}

// func (uc *Usecase) closeOrder(ctx context.Context, order *entities.Order) error {
// 	return uc.unholdBalance(ctx, order)
// }

// func (uc *Usecase) executeOrder(ctx context.Context, order *entities.Order) error {
// 	return uc.appendBalance(ctx, order)
// }

// func (uc *Usecase) validateOrder(ctx context.Context, order *entities.Order) error {
// 	account, err := uc.getAccountByUID(ctx, order.AccountUID)
// 	if err != nil {
// 		uc.log.Error(fmt.Sprintf("validateOrder:GetAccountByUID [AccountUID: %s] error: %v", order.AccountUID, err))
// 		return err
// 	}

// 	if account.PositionMode == entities.PositionModeOneway {
// 		order.PositionSide = entities.PositionSideBoth
// 	} else {
// 		if order.PositionSide != entities.PositionSideLong && order.PositionSide != entities.PositionSideShort {
// 			return apperror.ErrInvalidPositionSide
// 		}
// 	}

// 	return nil
// }

func (uc *Usecase) saveOrder(ctx context.Context, order *entities.Order) error {
	return uc.order.InsertOrder(ctx, order)
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
