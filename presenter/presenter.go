package presenter

import "github.com/otyang/waas-go"

type AllWalletsPresenter struct{}

func AllWallets(wallets []waas.Wallet, currencies []waas.Currency) (*AllWalletsPresenter, error) {
	if len(wallets) == 0 {
		return
	}
}
