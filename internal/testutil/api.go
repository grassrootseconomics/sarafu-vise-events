package testutil

import (
	"context"

	"git.defalsify.org/vise.git/logging"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
)

var (
	logg = logging.NewVanilla().WithDomain("sarafu-vise-events.testutil")
)

const (
	AliceChecksum = "0xeae046BF396e91f5A8D74f863dC57c107c8a4a70"
	BobChecksum = "0xB3117202371853e24B725d4169D87616A7dDb127"
	AliceSession = "5553425"
)

type MockApi struct {
	TransactionsContent []dataserviceapi.Last10TxResponse
	VouchersContent []dataserviceapi.TokenHoldings
	VoucherDataContent *models.VoucherDataResult
}

func(m MockApi) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	return nil, nil
}

func(m MockApi) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	return nil, nil
}

func(m MockApi) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	return nil, nil
}

func(m MockApi) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	logg.DebugCtxf(ctx, "mockapi fetchvouchers", "key", publicKey)
	return m.VouchersContent, nil
}

func(m MockApi) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	logg.DebugCtxf(ctx, "mockapi fetchtransactions", "key", publicKey)
	return m.TransactionsContent, nil
}

func(m MockApi) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	return m.VoucherDataContent, nil
}

func(m MockApi) CheckAliasAddress(ctx context.Context, alias string) (*dataserviceapi.AliasAddress, error) {
	return nil, nil
}

func(m MockApi) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	return nil, nil
}
