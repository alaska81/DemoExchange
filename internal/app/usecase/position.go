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

	positions, err := uc.position.SelectOpenPositions(ctx, accountUID)
	if err != nil {
		uc.log.Error(fmt.Sprintf("PositionsList:SelectOpenPositions [AccountUID: %s] error: %v", accountUID, err))
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
		ts := entities.TS()

		if account.PositionMode == entities.PositionModeOneway {
			position, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideBoth}]
			if !ok {
				position = entities.NewPosition(account, exchange, symbol, entities.PositionSideBoth)
			}
			position.MarkPrice = ticker.Last
			position.UpdateTS = ts
			result = append(result, position)
		} else {
			positionLong, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideLong}]
			if !ok {
				positionLong = entities.NewPosition(account, exchange, symbol, entities.PositionSideLong)
			}
			positionLong.MarkPrice = ticker.Last
			positionLong.UpdateTS = ts
			result = append(result, positionLong)

			positionShort, ok := mapPositions[key{symbol: symbol, side: entities.PositionSideShort}]
			if !ok {
				positionShort = entities.NewPosition(account, exchange, symbol, entities.PositionSideShort)
			}
			positionShort.MarkPrice = ticker.Last
			positionShort.UpdateTS = ts
			result = append(result, positionShort)
		}
	}

	return result, nil
}

func (uc *Usecase) SetPositionType(ctx context.Context, accountUID entities.AccountUID, positionType entities.PositionType) error {

	return nil
}

func (uc *Usecase) SetPositionLeverage(ctx context.Context, accountUID entities.AccountUID, leverage entities.PositionLeverage) error {

	return nil
}

func (uc *Usecase) getOpenPosition(ctx context.Context, tx pgx.Tx, order *entities.Order) (*entities.Position, error) {
	position, err := uc.position.SelectOpenPositionBySymbolSide(ctx, tx, order.AccountUID, order.Symbol, order.PositionSide)
	if err != nil {
		if !errors.Is(err, apperror.ErrPositionNotFound) {
			uc.log.Error(fmt.Sprintf("getPosition:SelectOpenPositionBySide [AccountUID: %s, position_side: %s] error: %v", order.AccountUID, order.PositionSide, err))
			return nil, err
		}

		account, err := uc.getAccountByUID(ctx, tx, order.AccountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("getPosition:getAccountByUID [AccountUID: %s] error: %v", order.AccountUID, err))
			return nil, err
		}

		ts := entities.TS()

		position = entities.NewPosition(account, order.Exchange, order.Symbol, order.PositionSide)
		position.Status = entities.PositionStatusOpen
		position.CreateTS = ts
		position.UpdateTS = ts
		position.IsNew = true
	}

	return position, nil
}

func (uc *Usecase) savePosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error {
	if position.IsNew {
		if err := uc.position.InsertPosition(ctx, tx, position); err != nil {
			uc.log.Error(fmt.Sprintf("setPosition:InsertPosition [%+v] error: %v", position, err))
			return err
		}
	}

	if err := uc.position.UpdatePosition(ctx, tx, position); err != nil {
		uc.log.Error(fmt.Sprintf("updatePosition:UpdatePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}
