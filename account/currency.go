package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
)

func (a *Account) CreateCurrency(ctx context.Context, currency types.Currency) (*types.Currency, error) {
	currency.CreatedAt = time.Now()
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(&currency).Exec(ctx)
	return &currency, err
}

func (a *Account) GetCurrency(ctx context.Context, currencyCode string) (*types.Currency, error) {
	currency := types.Currency{Code: currencyCode}
	err := a.db.NewSelect().Model(&currency).WherePK().Limit(1).Scan(ctx)
	return &currency, err
}

func (a *Account) UpdateCurrency(ctx context.Context, currency types.Currency) (*types.Currency, error) {
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(&currency).WherePK().Exec(ctx)
	return &currency, err
}

func (a *Account) ListCurrencies(ctx context.Context) ([]types.Currency, error) {
	var currencies []types.Currency
	err := a.db.NewSelect().Model(&currencies).Scan(ctx)
	return currencies, err
}

func (a *Account) DeleteCurrency(ctx context.Context, currencyCode string) error {
	_, err := a.db.NewDelete().Model(&types.Currency{Code: currencyCode}).WherePK().Exec(ctx)
	return err
}
