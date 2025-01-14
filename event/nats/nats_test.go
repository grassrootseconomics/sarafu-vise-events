package nats

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"git.defalsify.org/vise.git/db"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/config"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	storedb "git.grassecon.net/grassrootseconomics/sarafu-vise/store/db"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
	"git.grassecon.net/grassrootseconomics/common/hex"
	apimocks "git.grassecon.net/grassrootseconomics/sarafu-api/testutil/mocks"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/internal/testutil"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/application"
	viseevent "git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/event"
	"git.grassecon.net/grassrootseconomics/visedriver/testutil/mocks"
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

// TODO: jetstream, would have been nice of you to provide an easier way to make a mock msg
type testMsg struct {
	data []byte
}

func(m *testMsg) Ack() error {
	return nil
}

func(m *testMsg) Nak() error {
	return nil
}

func(m *testMsg) NakWithDelay(time.Duration) error {
	return nil
}

func(m *testMsg) Data() []byte {
	return m.data
}

func(m *testMsg) Reply() string {
	return ""
}

func(m *testMsg) Subject() string {
	return ""
}

func(m *testMsg) Term() error {
	return nil
}

func(m *testMsg) TermWithReason(string) error {
	return nil
}

func(m *testMsg) DoubleAck(ctx context.Context) error {
	return nil
}

func(m *testMsg) Headers() nats.Header {
	return nats.Header{}
}

func(m *testMsg) InProgress() error {
	return nil
}

func(m *testMsg) Metadata() (*jetstream.MsgMetadata, error) {
	return nil, nil
}

func TestHandleMsg(t *testing.T) {
	err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	api := &apimocks.MockApi{}
	api.TransactionsContent = []dataserviceapi.Last10TxResponse{
		dataserviceapi.Last10TxResponse{
			Sender: apimocks.AliceChecksum,
			Recipient: apimocks.BobChecksum,
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
	storageService := mocks.NewMemStorageService(ctx)
	eu := viseevent.NewEventsUpdater(api, storageService)
	userDb := storageService.Db

	alice, err := hex.NormalizeHex(apimocks.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, storedb.PackKey(storedb.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(apimocks.AliceSession))
	if err != nil {
		t.Fatal(err)
	}

	eh := eu.ToEventsHandler()
	sub := NewNatsSubscription(eh)
	_ = sub.Connect(ctx, "")

	data := fmt.Sprintf(`{
	"block": %d,
	"contractAddress": "%s",
	"success": true,
	"timestamp": %d,
	"transactionHash": "%s",
	"transactionType": "TOKEN_TRANSFER",
	"payload": {
		"from": "%s",
		"to": "%s",
		"value": "%d"
	}
}`, txBlock, tokenAddress, txTimestamp, txHash, apimocks.AliceChecksum, apimocks.BobChecksum, txValue)
	msg := &testMsg{
		data: []byte(data),
	}
	sub.handleEvent(msg)

	userStore := store.UserDataStore{
		Db: userDb,
	}
	v, err := userStore.ReadEntry(ctx, apimocks.AliceSession, storedb.DATA_ACTIVE_SYM)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte(tokenSymbol)) {
		t.Fatalf("expected '%s', got %s", tokenSymbol, v)
	}

	v, err = userStore.ReadEntry(ctx, apimocks.AliceSession, storedb.DATA_ACTIVE_BAL)
	if err != nil {
		t.Fatal(err)
	}
	fmts := fmt.Sprintf("%%1.%df", tokenDecimals)
	expect := fmt.Sprintf(fmts, float64(tokenBalance) / math.Pow(10, tokenDecimals))
	//if !bytes.Equal(v, []byte(strconv.Itoa(tokenBalance))) {
	if !bytes.Equal(v, []byte(expect)) {
		t.Fatalf("expected '%d', got %s", tokenBalance, v)
	}

	v, err = userStore.ReadEntry(ctx, apimocks.AliceSession, storedb.DATA_TRANSACTIONS)
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
	ctx = context.WithValue(ctx, "SessionId", apimocks.AliceSession)
	rrs, err := mh.GetVoucherList(ctx, "", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	expect = fmt.Sprintf("1:%s", tokenSymbol)
	if rrs.Content != expect {
		t.Fatalf("expected '%v', got '%v'", expect, rrs.Content)
	}
//	userDb.SetPrefix(event.DATATYPE_USERSUB)
//	userDb.SetSession(apimocks.AliceSession)
//	k := append([]byte("vouchers"), []byte("sym")...)
//	v, err = userDb.Get(ctx, k)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !bytes.Contains(v, []byte(fmt.Sprintf("1:%s", tokenSymbol))) {
//		t.Fatalf("expected '1:%s', got %s", tokenSymbol, v)
//	}
}
