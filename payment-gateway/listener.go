package payment_gateway

import (
	"context"
	"math/big"
	"strings"

	"payment-gateway/transport"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Listener struct {
	Provider *transport.Provider
	Portal   abi.ABI
	Payee    abi.ABI
	contract string
	Done     chan int
}

func NewListener(provider string, contract string) (*Listener, error) {

	client, err := transport.NewProvider(provider)
	if err != nil {
		log.Error("failed to open provider")
		return nil, err
	}

	payee, _ := abi.JSON(strings.NewReader(PayeeABI))
	portal, _ := abi.JSON(strings.NewReader(PortalABI))

	listener := &Listener{
		Provider: client,
		Payee:    payee,
		Portal:   portal,
		Done:     make(chan int, 1),
	}

	return listener, nil
}

func (l *Listener) ListenOnPayee(from int64, ch chan types.Log) {

	latest := l.Provider.GetLatestBlock()

	addresses := []common.Address{common.HexToAddress(l.contract)}

	ctx := context.Background()

	for {
		logs := l.Provider.FilterLogs(ctx, addresses, big.NewInt(from), big.NewInt(from+100))
		for _, event := range logs {
			ch <- event
		}

		if uint64(from+100) > latest {
			break
		}
	}

	l.Provider.ListenEvent(addresses, ch, big.NewInt(int64(latest)), l.Done)
}
