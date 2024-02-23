package usecase

import (
	"DemoExchange/internal/app/entities"
	"context"
	"fmt"
)

func (uc *Usecase) AppendTransaction(ctx context.Context, transaction *entities.Transaction) error {
	return uc.transaction.InsertTransaction(ctx, transaction)
}

func (uc *Usecase) TransactionsList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, filter entities.TransactionFilter) ([]*entities.Transaction, error) {
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	transaction, err := uc.transaction.SelectAccountTransactions(ctx, exchange, accountUID, filter)
	if err != nil {
		uc.log.Error(fmt.Sprintf("TransactionsList:SelectAccountTransactions [AccountUID: %s] error: %v", accountUID, err))
		return nil, err
	}

	return transaction, nil
}
