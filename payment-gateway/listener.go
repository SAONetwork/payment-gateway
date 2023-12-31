package payment_gateway

import (
	"context"
	"math/big"
	"payment-gateway/ethprovider"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Listener struct {
	Provider *ethprovider.Provider
	Portal   abi.ABI
	Payee    abi.ABI
	contract string
	Done     chan int
}

func NewListener(provider string, contract string) (*Listener, error) {

	client, err := ethprovider.NewProvider(provider)
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
		contract: contract,
		Done:     make(chan int, 1),
	}

	return listener, nil
}

func (l *Listener) ListenOnPayee(from int64, ch chan types.Log) {

	latest := l.Provider.GetLatestBlock()

	log.Infof("listen envent on %s from %d\n", l.contract, from)

	addresses := []common.Address{common.HexToAddress(l.contract)}

	ctx := context.Background()

	for {
		logs := l.Provider.FilterLogs(ctx, addresses, big.NewInt(from), big.NewInt(from+100))
		log.Infof("events %d %d %d", len(logs), from, from+100)
		for _, event := range logs {
			ch <- event
		}

		if len(logs) == 0 {
			ch <- types.Log{
				BlockNumber: uint64(from + 100),
			}
		}

		if uint64(from+100) > latest {
			from = int64(latest)
			break
		}
		from += 100
	}

	l.Provider.ListenEvent(addresses, ch, big.NewInt(int64(latest)), l.Done)
}
