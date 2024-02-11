package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
)

func (uc *Usecase) PositionsList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error) {
	account, err := uc.getAccountByUID(ctx, nil, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:getAccountByUID [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	r := tickers.New()
	tickers, err := r.GetTickers(exchange.Name())
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

		if account.PositionMode == entities.PositionModeOneway {
			position, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideBoth}]
			if !ok {
				position = entities.NewPosition(account, exchange, symbol, entities.PositionSideBoth)
			}
			position.MarkPrice = ticker.Last
			result = append(result, position)
		} else {
			positionLong, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideLong}]
			if !ok {
				positionLong = entities.NewPosition(account, exchange, symbol, entities.PositionSideLong)
			}
			positionLong.MarkPrice = ticker.Last
			result = append(result, positionLong)

			positionShort, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideShort}]
			if !ok {
				positionShort = entities.NewPosition(account, exchange, symbol, entities.PositionSideShort)
			}
			positionShort.MarkPrice = ticker.Last
			result = append(result, positionShort)
		}
	}

	return result, nil
}

func (uc *Usecase) SetPositionMarginType(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, marginType entities.MarginType) error {
	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, &symbol); err != nil {
			return apperror.ErrSetMarginType.Wrap(err)
		}

		positions, err := uc.getPositionsBySymbol(ctx, tx, exchange, accountUID, symbol)
		if err != nil {
			return apperror.ErrSetMarginType.Wrap(err)
		}

		for _, position := range positions {
			position := position
			if position.Amount != 0 {
				return apperror.ErrSetMarginType.Wrap(apperror.ErrPositionExists)
			}

			position.MarginType = marginType

			if err := uc.savePosition(ctx, tx, position); err != nil {
				return apperror.ErrSetMarginType.Wrap(err)
			}
		}

		return nil
	})
}

func (uc *Usecase) SetPositionLeverage(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, leverage entities.PositionLeverage) error {
	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, &symbol); err != nil {
			return apperror.ErrSetLeverage.Wrap(err)
		}

		positions, err := uc.getPositionsBySymbol(ctx, tx, exchange, accountUID, symbol)
		if err != nil {
			return apperror.ErrSetLeverage.Wrap(err)
		}

		for _, position := range positions {
			position := position
			if position.Amount != 0 {
				return apperror.ErrSetLeverage.Wrap(apperror.ErrPositionExists)
			}

			position.Leverage = leverage

			if err := uc.savePosition(ctx, tx, position); err != nil {
				return apperror.ErrSetLeverage.Wrap(err)
			}
		}

		return nil
	})
}

func (uc *Usecase) getPositionBySide(ctx context.Context, tx pgx.Tx, order *entities.Order) (*entities.Position, error) {
	position, err := uc.position.SelectPositionBySide(ctx, tx, order.AccountUID, order.Symbol, order.PositionSide)
	if err != nil {
		if !errors.Is(err, apperror.ErrPositionNotFound) {
			uc.log.Error(fmt.Sprintf("getPositionBySide:SelectPositionBySide [AccountUID: %s, position_side: %s] error: %v", order.AccountUID, order.PositionSide, err))
			return nil, err
		}

		account, err := uc.getAccountByUID(ctx, tx, order.AccountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("getPositionBySide:getAccountByUID [AccountUID: %s] error: %v", order.AccountUID, err))
			return nil, err
		}

		position = entities.NewPosition(account, order.Exchange, order.Symbol, order.PositionSide)
	}

	return position, nil
}

func (uc *Usecase) getPositionsBySymbol(ctx context.Context, tx pgx.Tx, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol) (map[entities.PositionSide]*entities.Position, error) {
	account, err := uc.getAccountByUID(ctx, nil, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("getPositionBySymbol:getAccountByUID [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	positions, err := uc.position.SelectPositionsBySymbol(ctx, tx, accountUID, symbol)
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

func (uc *Usecase) savePosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error {
	if position.IsNew {
		if err := uc.position.InsertPosition(ctx, tx, position); err != nil {
			uc.log.Error(fmt.Sprintf("savePosition:InsertPosition [%+v] error: %v", position, err))
			return err
		}
	}

	if err := uc.position.UpdatePosition(ctx, tx, position); err != nil {
		uc.log.Error(fmt.Sprintf("savePosition:UpdatePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}
