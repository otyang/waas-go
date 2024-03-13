package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go/currency"
)

func (a *Account) CreateCurrency(ctx context.Context, currency currency.Currency) (*currency.Currency, error) {
	currency.CreatedAt = time.Now()
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(&currency).Exec(ctx)
	return &currency, err
}

func (a *Account) UpdateCurrency(ctx context.Context, currency currency.Currency) (*currency.Currency, error) {
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(&currency).WherePK().Exec(ctx)
	return &currency, err
}

func (a *Account) ListCurrencies(ctx context.Context) ([]currency.Currency, error) {
	var currencies []currency.Currency

	err := a.db.NewSelect().Model(&currencies).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return currencies, nil
}
