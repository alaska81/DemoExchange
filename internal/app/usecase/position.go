package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

const processTimeout = 10 * time.Second

var muPositionProcess = sync.Mutex{}

func (uc *Usecase) PositionsList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error) {
	account, err := uc.GetAccountByUID(ctx, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:GetAccountByUID [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	balances, err := uc.GetBalances(ctx, exchange, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:GetBalances [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	tickers, err := uc.tickers.GetTickers(exchange.Name())
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:GetTickers [exchange: %s] error: %v", exchange.Name(), err))
		return nil, err
	}

	count := len(tickers)
	if account.PositionMode == entities.PositionModeHedge {
		count *= 2
	}

	result := make([]*entities.Position, 0, count)

	positions, err := uc.position.SelectAccountPositions(ctx, exchange, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:SelectAccountPositions [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	type key struct {
		symbol entities.Symbol
		side   entities.PositionSide
	}

	mapPositions := make(map[key]*entities.Position, len(positions))
	for _, position := range positions {
		mapPositions[key{symbol: position.Symbol, side: position.Side}] = position
	}

	for _, ticker := range tickers {
		symbol := entities.Symbol(ticker.Symbol)
		coins := symbol.GetCoins()
		balance := balances[coins.CoinBase].WalletBalance

		if account.PositionMode == entities.PositionModeOneway {
			position, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideBoth}]
			if !ok {
				position = entities.NewPosition(account, exchange, symbol, entities.PositionSideBoth)
			}

			position.CalcMarginBalance(ticker.Last)
			position.CalcLiquidationPrice(balance)
			result = append(result, position)
		} else {
			positionLong, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideLong}]
			if !ok {
				positionLong = entities.NewPosition(account, exchange, symbol, entities.PositionSideLong)
			}

			positionLong.CalcMarginBalance(ticker.Last)
			positionLong.CalcLiquidationPrice(balance)
			result = append(result, positionLong)

			positionShort, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideShort}]
			if !ok {
				positionShort = entities.NewPosition(account, exchange, symbol, entities.PositionSideShort)
			}

			positionShort.CalcMarginBalance(ticker.Last)
			positionShort.CalcLiquidationPrice(balance)
			result = append(result, positionShort)
		}
	}

	return result, nil
}

func (uc *Usecase) SetPositionMarginType(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, marginType entities.MarginType) error {
	return uc.position.WithTx(ctx, func(ctx context.Context) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, &symbol); err != nil {
			return apperror.ErrSetMarginType.Wrap(err)
		}

		positions, err := uc.getPositionsBySymbol(ctx, exchange, accountUID, symbol)
		if err != nil {
			return apperror.ErrSetMarginType.Wrap(err)
		}

		for _, position := range positions {
			position := position
			if position.Amount != 0 {
				return apperror.ErrSetMarginType.Wrap(apperror.ErrPositionExists)
			}

			position.MarginType = marginType

			if err := uc.SavePosition(ctx, position); err != nil {
				return apperror.ErrSetMarginType.Wrap(err)
			}
		}

		return nil
	})
}

func (uc *Usecase) SetPositionLeverage(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, leverage entities.PositionLeverage) error {
	return uc.position.WithTx(ctx, func(ctx context.Context) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, &symbol); err != nil {
			return apperror.ErrSetLeverage.Wrap(err)
		}

		positions, err := uc.getPositionsBySymbol(ctx, exchange, accountUID, symbol)
		if err != nil {
			return apperror.ErrSetLeverage.Wrap(err)
		}

		for _, position := range positions {
			position := position
			if position.Amount != 0 {
				return apperror.ErrSetLeverage.Wrap(apperror.ErrPositionExists)
			}

			position.Leverage = leverage

			if err := uc.SavePosition(ctx, position); err != nil {
				return apperror.ErrSetLeverage.Wrap(err)
			}
		}

		return nil
	})
}

func (uc *Usecase) getPositionsBySymbol(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol) (map[entities.PositionSide]*entities.Position, error) {
	account, err := uc.GetAccountByUID(ctx, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("getPositionBySymbol:GetAccountByUID [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	positions, err := uc.position.SelectPositionsBySymbol(ctx, accountUID, symbol)
	if err != nil {
		uc.log.Error(fmt.Sprintf("getPositionBySymbol:SelectPositionsBySymbol [AccountUID: %s, symbol: %s] error: %v", accountUID, symbol, err))
		return nil, err
	}

	if account.PositionMode == entities.PositionModeOneway {
		_, ok := positions[entities.PositionSideBoth]
		if !ok {
			positions[entities.PositionSideBoth] = entities.NewPosition(account, exchange, symbol, entities.PositionSideBoth)
		}
	} else {
		_, ok := positions[entities.PositionSideLong]
		if !ok {
			positions[entities.PositionSideLong] = entities.NewPosition(account, exchange, symbol, entities.PositionSideLong)
		}

		_, ok = positions[entities.PositionSideShort]
		if !ok {
			positions[entities.PositionSideShort] = entities.NewPosition(account, exchange, symbol, entities.PositionSideShort)
		}
	}

	return positions, nil
}

func (uc *Usecase) checkPresentOpenPosition(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) error {
	positions, err := uc.position.SelectAccountOpenPositions(ctx, exchange, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("checkPresentOpenPosition:SelectAccountOpenPositions [account_uid: %v] error: %v", accountUID, err))
		return apperror.ErrRequestError
	}

	if len(positions) > 0 {
		return apperror.ErrPositionExists
	}

	return nil
}

func (uc *Usecase) GetPositionBySide(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error) {
	position, err := uc.position.SelectPositionBySide(ctx, accountUID, symbol, side)
	if err != nil {
		if !errors.Is(err, apperror.ErrPositionNotFound) {
			uc.log.Error(fmt.Sprintf("getPositionBySide:SelectPositionBySide [AccountUID: %s, position_side: %s] error: %v", accountUID, side, err))
			return nil, err
		}

		account, err := uc.GetAccountByUID(ctx, accountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("getPositionBySide:GetAccountByUID [AccountUID: %s] error: %v", accountUID, err))
			return nil, err
		}

		position = entities.NewPosition(account, exchange, symbol, side)
	}

	return position, nil
}

func (uc *Usecase) SavePosition(ctx context.Context, position *entities.Position) error {
	if position.IsNew {
		if err := uc.position.InsertPosition(ctx, position); err != nil {
			uc.log.Error(fmt.Sprintf("savePosition:InsertPosition [%+v] error: %v", *position, err))
			return err
		}

		uc.chPositions <- position
		return nil
	}

	uc.log.Info(fmt.Sprintf("savePosition:UpdatePosition [%+v]", *position))

	if err := uc.position.UpdatePosition(ctx, position); err != nil {
		uc.log.Error(fmt.Sprintf("savePosition:UpdatePosition [%+v] error: %v", *position, err))
		return err
	}

	uc.log.Info(fmt.Sprintf("savePosition:UpdatePosition222 [%+v]", *position))

	uc.chPositions <- position
	return nil
}

func (uc *Usecase) updatePosition(ctx context.Context, position *entities.Position) error {
	ctx, cancel := context.WithTimeout(ctx, TimeoutUpdate)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			uc.log.Error(fmt.Sprintf("updatePosition:Done [%+v] error: %v", *position, ctx.Err()))
			return ctx.Err()
		default:
			position.UpdateTS = entities.TS()
			if err := uc.position.UpdatePosition(ctx, position); err != nil {
				uc.log.Error(fmt.Sprintf("updatePosition:UpdatePosition [%+v] error: %v", *position, err))
				time.Sleep(TimeoutRetry)
				continue
			}
			return nil
		}
	}
}

func (uc *Usecase) ProcessPositions(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			uc.log.Error(fmt.Sprintf("ProcessPositions:Done error: %v", ctx.Err()))
			return

		case position := <-uc.chPositions:
			uc.log.Info(fmt.Sprintf("ProcessPositions [%+v]", *position))

			go func(position *entities.Position) {
				muPositionProcess.Lock()
				old, ok := uc.cachePositions.Get(position.PositionUID)
				if ok {
					old.Amount = position.Amount
					old.Price = position.Price
					old.Margin = position.Margin
					old.HoldAmount = position.HoldAmount
					old.UpdateTS = position.UpdateTS
					muPositionProcess.Unlock()
					return
				}
				muPositionProcess.Unlock()

				uc.cachePositions.Set(position.PositionUID, position)
				defer func() {
					uc.cachePositions.Delete(position.PositionUID)
					uc.log.Info(fmt.Sprintf("ProcessPositions:Delete [%+v]", *position))
				}()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						muPositionProcess.Lock()
						if position.Amount == 0 {
							uc.log.Info(fmt.Sprintf("ProcessPositions:Amount=0 [%+v]", *position))
							muPositionProcess.Unlock()
							return
						}

						if uc.checkPositionLiquidation(ctx, position) {
							uc.log.Info(fmt.Sprintf("ProcessPositions:checkPositionLiquidation [%+v]", *position))
							muPositionProcess.Unlock()
							return
						}

						muPositionProcess.Unlock()
						time.Sleep(processTimeout)
					}
				}
			}(position)
		}
	}
}

func (uc *Usecase) ProcessOpenPositions(ctx context.Context) error {
	positions, err := uc.position.SelectOpenPositions(ctx)
	if err != nil {
		uc.log.Error(fmt.Sprintf("ProcessOpenPositions:SelectOpenPositions error: %v", err))
		return err
	}

	uc.log.Info("ProcessOpenPositions: ", len(positions))

	for _, position := range positions {
		position := position
		go func() {
			uc.chPositions <- position
		}()

	}

	return nil
}

func (uc *Usecase) checkPositionLiquidation(ctx context.Context, position *entities.Position) bool {
	ticker, err := uc.tickers.GetTickerWithContext(ctx, position.Exchange.Name(), position.Symbol.String())
	if err != nil {
		return false
	}

	position.CalcMarginBalance(ticker.Last)

	if position.MarginBalance <= 0 {
		err := uc.position.WithTx(ctx, func(ctx context.Context) error {
			position.Amount = 0
			position.HoldAmount = 0
			position.Margin = 0
			if err := uc.updatePosition(ctx, position); err != nil {
				uc.log.Error(fmt.Sprintf("checkPositionLiquidation:updatePosition [%+v] error: %v", *position, err))
				return err
			}

			transaction := entities.NewTransaction(position.AccountUID, position.Exchange, position.Symbol, entities.TransactionTypeLiquidation, position.Amount)
			if err := uc.AppendTransaction(ctx, transaction); err != nil {
				uc.log.Error(fmt.Sprintf("checkPositionLiquidation:AppendTransaction [%+v] error: %v", *transaction, err))
				return err
			}

			return nil
		})

		return err == nil
	}

	return false
}
