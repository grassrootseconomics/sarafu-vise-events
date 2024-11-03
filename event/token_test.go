package event

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"git.defalsify.org/vise.git/db"
	memdb "git.defalsify.org/vise.git/db/mem"
	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/models"
	"git.grassecon.net/term/lookup"
	"git.grassecon.net/urdt/ussd/common"
	"git.grassecon.net/term/internal/testutil"
)

const (
	txBlock = 42
	tokenAddress = "0x765DE816845861e75A25fCA122bb6898B8B1282a"
	tokenSymbol = "FOO"
	tokenName = "Foo Token"
	tokenDecimals = 6
	txValue = 1337
	tokenBalance = 362436
	txTimestamp = 1730592500
	txHash = "0xabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	sinkAddress = "0xb42C5920014eE152F2225285219407938469BBfA"
)

func TestTokenTransfer(t *testing.T) {
	err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	api := &testutil.MockApi{}
	api.TransactionsContent = []dataserviceapi.Last10TxResponse{
		dataserviceapi.Last10TxResponse{
			Sender: testutil.AliceChecksum,
			Recipient: testutil.BobChecksum,
			TransferValue: strconv.Itoa(txValue),
			ContractAddress: tokenAddress,
			TxHash: txHash,
			DateBlock: time.Unix(txTimestamp, 0),
			TokenSymbol: tokenSymbol,
			TokenDecimals: strconv.Itoa(tokenDecimals),
		},
	}
	api.VoucherDataContent = &models.VoucherDataResult{
		TokenSymbol: tokenSymbol,
		TokenName: tokenName,
		TokenDecimals: strconv.Itoa(tokenDecimals),
		SinkAddress: sinkAddress,
	}
	api.VouchersContent = []dataserviceapi.TokenHoldings{
		dataserviceapi.TokenHoldings{
			ContractAddress: tokenAddress,
			TokenSymbol: tokenSymbol,
			TokenDecimals: strconv.Itoa(tokenDecimals),
			Balance: strconv.Itoa(tokenBalance),
		},
	}
	lookup.Api = api

	ctx := context.Background()
	userDb := memdb.NewMemDb()
	err = userDb.Connect(ctx, "")
	if err != nil {
		panic(err)
	}

	alice, err := common.NormalizeHex(testutil.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: deduplicate test setup
	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, common.PackKey(common.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(testutil.AliceSession))
	if err != nil {
		t.Fatal(err)
	}
	store := common.UserDataStore{
		Db: userDb,
	}

	ev := &eventTokenTransfer{
		From: testutil.BobChecksum,
		To: testutil.AliceChecksum,
		Value: txValue,
	}
	err = handleTokenTransfer(ctx, &store, ev)
	if err != nil {
		t.Fatal(err)
	}

	v, err := store.ReadEntry(ctx, testutil.AliceSession, common.DATA_ACTIVE_SYM)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(tokenSymbol)) {
		t.Fatalf("expected '%s', got %s", tokenSymbol, v)
	}

	v, err = store.ReadEntry(ctx, testutil.AliceSession, common.DATA_ACTIVE_BAL)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(strconv.Itoa(tokenBalance))) {
		t.Fatalf("expected '%d', got %s", tokenBalance, v)
	}

	v, err = store.ReadEntry(ctx, testutil.AliceSession, common.DATA_TRANSACTIONS)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte("abcdef")) {
		t.Fatal("no transaction data")
	}

	userDb.SetPrefix(DATATYPE_USERSUB)
	userDb.SetSession(testutil.AliceSession)
	k := append([]byte("vouchers"), []byte("sym")...)
	v, err = userDb.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte(fmt.Sprintf("1:%s", tokenSymbol))) {
		t.Fatalf("expected '1:%s', got %s", tokenSymbol, v)
	}


}

func TestTokenMint(t *testing.T) {
	err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	api := &testutil.MockApi{}
	api.TransactionsContent = []dataserviceapi.Last10TxResponse{
		dataserviceapi.Last10TxResponse{
			Sender: testutil.AliceChecksum,
			Recipient: testutil.BobChecksum,
			TransferValue: strconv.Itoa(txValue),
			ContractAddress: tokenAddress,
			TxHash: txHash,
			DateBlock: time.Unix(txTimestamp, 0),
			TokenSymbol: tokenSymbol,
			TokenDecimals: strconv.Itoa(tokenDecimals),
		},
	}
	api.VoucherDataContent = &models.VoucherDataResult{
		TokenSymbol: tokenSymbol,
		TokenName: tokenName,
		TokenDecimals: strconv.Itoa(tokenDecimals),
		SinkAddress: sinkAddress,
	}
	api.VouchersContent = []dataserviceapi.TokenHoldings{
		dataserviceapi.TokenHoldings{
			ContractAddress: tokenAddress,
			TokenSymbol: tokenSymbol,
			TokenDecimals: strconv.Itoa(tokenDecimals),
			Balance: strconv.Itoa(tokenBalance),
		},
	}
	lookup.Api = api

	ctx := context.Background()
	userDb := memdb.NewMemDb()
	err = userDb.Connect(ctx, "")
	if err != nil {
		panic(err)
	}

	alice, err := common.NormalizeHex(testutil.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, common.PackKey(common.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(testutil.AliceSession))
	if err != nil {
		t.Fatal(err)
	}
	store := common.UserDataStore{
		Db: userDb,
	}

	ev := &eventTokenMint{
		To: testutil.AliceChecksum,
		Value: txValue,
	}
	err = handleTokenMint(ctx, &store, ev)
	if err != nil {
		t.Fatal(err)
	}

	v, err := store.ReadEntry(ctx, testutil.AliceSession, common.DATA_ACTIVE_SYM)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(tokenSymbol)) {
		t.Fatalf("expected '%s', got %s", tokenSymbol, v)
	}

	v, err = store.ReadEntry(ctx, testutil.AliceSession, common.DATA_ACTIVE_BAL)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(strconv.Itoa(tokenBalance))) {
		t.Fatalf("expected '%d', got %s", tokenBalance, v)
	}

	v, err = store.ReadEntry(ctx, testutil.AliceSession, common.DATA_TRANSACTIONS)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte("abcdef")) {
		t.Fatal("no transaction data")
	}

	userDb.SetPrefix(DATATYPE_USERSUB)
	userDb.SetSession(testutil.AliceSession)
	k := append([]byte("vouchers"), []byte("sym")...)
	v, err = userDb.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte(fmt.Sprintf("1:%s", tokenSymbol))) {
		t.Fatalf("expected '1:%s', got %s", tokenSymbol, v)
	}

}
