package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go"
)

func (a *Account) CreateCurrency(ctx context.Context, currency waas.Currency) (*waas.Currency, error) {
	currency.CreatedAt = time.Now()
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(&currency).Exec(ctx)
	return &currency, err
}

func (a *Account) GetCurrency(ctx context.Context, currencyCode string) (*waas.Currency, error) {
	currency := waas.Currency{Code: currencyCode}
	err := a.db.NewSelect().Model(&currency).WherePK().Limit(1).Scan(ctx)
	return &currency, err
}

func (a *Account) UpdateCurrency(ctx context.Context, currency waas.Currency) (*waas.Currency, error) {
	currency.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(&currency).WherePK().Exec(ctx)
	return &currency, err
}

func (a *Account) ListCurrencies(ctx context.Context) ([]waas.Currency, error) {
	var currencies []waas.Currency
	err := a.db.NewSelect().Model(&currencies).Scan(ctx)
	return currencies, err
}

func (a *Account) DeleteCurrency(ctx context.Context, currencyCode string) error {
	_, err := a.db.NewDelete().Model(&waas.Currency{Code: currencyCode}).WherePK().Exec(ctx)
	return err
}
