package waas

import (
	"github.com/otyang/waas-go/account"
	"github.com/otyang/waas-go/presenter"
	"github.com/uptrace/bun"
)

type Client struct {
	Account   *account.Client
	Presenter *presenter.Client
}

func New(db *bun.DB) *Client {
	return &Client{
		Account:   account.New(db),
		Presenter: presenter.New(),
	}
}

func NewWithMigration(db *bun.DB) (*Client, error) {
	acc, err := account.NewWithMigration(db)
	if err != nil {
		return nil, err
	}
	return &Client{
		Account:   acc,
		Presenter: presenter.New(),
	}, nil
}
