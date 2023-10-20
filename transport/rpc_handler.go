package transport

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/SaoNetwork/sao-node/node/config"
	"github.com/SaoNetwork/sao-node/types"
	"github.com/ipfs/go-datastore"
	"payment-gateway/api"
)

type RpcHandler struct {
	Ctx              context.Context
	DbLk             sync.Mutex
	Db               datastore.Batching
	GatewayApi       api.SaoApi
	StagingPath      string
	StagingSapceSize int64
}

func NewHandler(ctx context.Context, ga api.SaoApi, db datastore.Batching, cfg *config.Node, stagingPath string) *RpcHandler {

	handler := RpcHandler{
		Ctx:              ctx,
		Db:               db,
		GatewayApi:       ga,
		StagingPath:      stagingPath,
		StagingSapceSize: cfg.Transport.StagingSapceSize,
	}
	return &handler
}

func (rs *RpcHandler) StoreProposal(params []string) (string, error) {
	if len(params) != 4 {
		return "", types.Wrapf(types.ErrInvalidParameters, "invalid params length")
	}

	var orderProposal types.OrderStoreProposal
	err := json.Unmarshal([]byte(params[0]), &orderProposal)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}

	res, err := rs.GatewayApi.StoreProposal(rs.Ctx, orderProposal)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}
	return res, nil
}
