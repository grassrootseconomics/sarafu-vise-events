package event

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"git.defalsify.org/vise.git/db"
	memdb "git.defalsify.org/vise.git/db/mem"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/config"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/application"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	storedb "git.grassecon.net/grassrootseconomics/sarafu-vise/store/db"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
	"git.grassecon.net/grassrootseconomics/common/hex"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/internal/testutil"
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
	bogusSym = "/-21380u"
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
		//TokenDecimals: strconv.Itoa(tokenDecimals),
		TokenDecimals: tokenDecimals,
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

	alice, err := hex.NormalizeHex(testutil.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: deduplicate test setup
	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, storedb.PackKey(storedb.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(testutil.AliceSession))
	if err != nil {
		t.Fatal(err)
	}
	userStore := store.UserDataStore{
		Db: userDb,
	}

	ev := &eventTokenTransfer{
		From: testutil.BobChecksum,
		To: testutil.AliceChecksum,
		Value: txValue,
	}
	err = handleTokenTransfer(ctx, &userStore, ev)
	if err != nil {
		t.Fatal(err)
	}

	v, err := userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_ACTIVE_SYM)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(tokenSymbol)) {
		t.Fatalf("expected '%s', got %s", tokenSymbol, v)
	}

	v, err = userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_ACTIVE_BAL)
	if err != nil {
		t.Fatal(err)
	}
	//if !bytes.Equal(v, []byte(strconv.Itoa(tokenBalance))) {
	fmts := fmt.Sprintf("%%1.%df", tokenDecimals)
	expect := fmt.Sprintf(fmts, float64(tokenBalance) / math.Pow(10, tokenDecimals))
	if !bytes.Equal(v, []byte(expect)) {
		t.Fatalf("expected '%s', got %s", expect, v)
	}

	v, err = userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_TRANSACTIONS)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte("abcdef")) {
		t.Fatal("no transaction data")
	}

	mh, err := application.NewMenuHandlers(nil, userStore, nil, nil, testutil.ReplaceSeparatorFunc)
	if err != nil {
		t.Fatal(err)
	}
	ctx = context.WithValue(ctx, "SessionId", testutil.AliceSession)
	rrs, err := mh.GetVoucherList(ctx, "", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	expect = fmt.Sprintf("1:%s", tokenSymbol)
	if rrs.Content != expect {
		t.Fatalf("expected '%v', got '%v'", expect, rrs.Content)
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
		//TokenDecimals: strconv.Itoa(tokenDecimals),
		TokenDecimals: tokenDecimals,
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

	alice, err := hex.NormalizeHex(testutil.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, storedb.PackKey(storedb.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(testutil.AliceSession))
	if err != nil {
		t.Fatal(err)
	}
	userStore := store.UserDataStore{
		Db: userDb,
	}

	ev := &eventTokenMint{
		To: testutil.AliceChecksum,
		Value: txValue,
	}
	err = handleTokenMint(ctx, &userStore, ev)
	if err != nil {
		t.Fatal(err)
	}

	v, err := userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_ACTIVE_SYM)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(tokenSymbol)) {
		t.Fatalf("expected '%s', got %s", tokenSymbol, v)
	}

	v, err = userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_ACTIVE_BAL)
	if err != nil {
		t.Fatal(err)
	}
	fmts := fmt.Sprintf("%%1.%df", tokenDecimals)
	expect := fmt.Sprintf(fmts, float64(tokenBalance) / math.Pow(10, tokenDecimals))
	//if !bytes.Equal(v, []byte(strconv.Itoa(tokenBalance))) {
	if !bytes.Equal(v, []byte(expect)) {
		t.Fatalf("expected '%d', got %s", tokenBalance, v)
	}

	v, err = userStore.ReadEntry(ctx, testutil.AliceSession, storedb.DATA_TRANSACTIONS)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(v, []byte("abcdef")) {
		t.Fatal("no transaction data")
	}


	mh, err := application.NewMenuHandlers(nil, userStore, nil, nil, testutil.ReplaceSeparatorFunc)
	if err != nil {
		t.Fatal(err)
	}
	ctx = context.WithValue(ctx, "SessionId", testutil.AliceSession)
	rrs, err := mh.GetVoucherList(ctx, "", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	expect = fmt.Sprintf("1:%s", tokenSymbol)
	if rrs.Content != expect {
		t.Fatalf("expected '%v', got '%v'", expect, rrs.Content)
	}
}
