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
	TimeoutUpdate     = 60 * time.Minute
	TimeoutRetry      = 10 * time.Second
	Limit         int = 100
)

var muOrder = &sync.RWMutex{}

func (uc *Usecase) NewOrder(ctx context.Context, o *entities.Order) error {
	var err error

	if o.Exchange == entities.ExchangeFutures {
		account, err := uc.GetAccountByUID(ctx, o.AccountUID)
		if err != nil {
			return err
		}
		o.PositionMode = account.PositionMode
	}

	order, err := orders.NewOrder(ctx, uc.markets, o)
	if err != nil {
		return err
	}

	if err := order.Validate(ctx); err != nil {
		uc.log.Error(fmt.Sprintf("NewOrder:Validate [%+v] error: %v", *order.GetOrder(), err))
		return err
	}

	if err := uc.order.WithTx(ctx, func(ctx context.Context) error {
		muOrder.Lock()
		defer muOrder.Unlock()

		if err := order.HoldBalance(ctx, uc, uc.log); err != nil {
			uc.log.Error(fmt.Sprintf("NewOrder:HoldBalance [%+v] error: %v", *order.GetOrder(), err))
			return err
		}

		if err := uc.saveOrder(ctx, order.GetOrder()); err != nil {
			uc.log.Error(fmt.Sprintf("NewOrder:saveOrder [%+v] error: %v", *order.GetOrder(), err))
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	go func() {
		uc.chOrders <- order
	}()

	return nil
}

func (uc *Usecase) ProcessOrders(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			uc.log.Error(fmt.Sprintf("ProcessOrders:Done error: %v", ctx.Err()))
			return

		case order := <-uc.chOrders:

			go func(order *orders.Order) {
				o := order.GetOrder()
				uc.log.Info(fmt.Sprintf("ProcessOrders [%+v]", *o))

				uc.cacheOrders.Set(o.OrderUID, o)
				defer func() {
					uc.cacheOrders.Delete(o.OrderUID)
					uc.log.Info(fmt.Sprintf("ProcessOrders:Process:close [%+v]", *o))
				}()

				ch, err := order.Process(ctx)
				if err != nil {
					uc.log.Error(fmt.Sprintf("ProcessOrders:Process [%+v] error: %v", *o, err))
					return
				}

				for {
					select {
					case <-ctx.Done():
						return
					case status, ok := <-ch:
						uc.log.Info(fmt.Sprintf("ProcessOrders:channel [%+v]: %v", status, ok))
						if !ok {
							return
						}

						muOrder.Lock()

						if status == entities.OrderStatusSuccess {
							uc.log.Info(fmt.Sprintf("ProcessOrders:AppendBalance [%+v]", *o))
							if err := order.AppendBalance(ctx, uc, uc.log); err != nil {
								uc.log.Error(fmt.Sprintf("ProcessOrders:AppendBalance [%+v] error: %v", *o, err))
								status = entities.OrderStatusFailed
								o.Error = err.Error()
							}
						}

						if status == entities.OrderStatusCancelled || status == entities.OrderStatusFailed {
							uc.log.Info(fmt.Sprintf("ProcessOrders:UnholdBalance [%+v]", *o))
							if err := order.UnholdBalance(ctx, uc, uc.log); err != nil {
								uc.log.Error(fmt.Sprintf("ProcessOrders:UnholdBalance [%+v] error: %v", *o, err))
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
			}(order)
		}
	}
}

func (uc *Usecase) ProcessPendingOrders(ctx context.Context) error {
	pendingOrders, err := uc.order.SelectPendingOrders(ctx)
	if err != nil {
		uc.log.Error(fmt.Sprintf("ProcessPendingOrders:SelectPendingOrders error: %v", err))
		return err
	}

	uc.log.Info("ProcessPendingOrders: ", len(pendingOrders))

	for _, o := range pendingOrders {
		if o.Exchange == entities.ExchangeFutures {
			o.PositionMode = o.PositionSide.PositionMode()
		}

		order, err := orders.NewOrder(ctx, uc.markets, o)
		if err != nil {
			return err
		}

		go func() {
			uc.chOrders <- order
		}()

	}

	return nil
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
	order, ok := uc.cacheOrders.Get(orderUID)
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

func (uc *Usecase) saveOrder(ctx context.Context, order *entities.Order) error {
	return uc.order.InsertOrder(ctx, order)
}

func (uc *Usecase) updateOrder(ctx context.Context, order *entities.Order) {
	ctx, cancel := context.WithTimeout(ctx, TimeoutUpdate)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			uc.log.Error(fmt.Sprintf("updateOrder:Done [%+v] error: %v", *order, ctx.Err()))
			return
		default:
			uc.log.Info(fmt.Sprintf("updateOrder [%+v]", *order))
			order.UpdateTS = entities.TS()
			if err := uc.order.UpdateOrder(ctx, order); err != nil {
				uc.log.Error(fmt.Sprintf("updateOrder:UpdateOrder [%+v] error: %v", *order, err))
				time.Sleep(TimeoutRetry)
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
