package api

import (
	"context"
	types2 "payment-gateway/types"

	"github.com/SaoNetwork/sao-node/types"
	"github.com/filecoin-project/go-jsonrpc/auth"
)

type SaoApi interface {
	// MethodGroup: Auth

	AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) //perm:none
	AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error)    //perm:admin

	// GetNodeAddress get current node's sao chain address
	GetNodeAddress(ctx context.Context) (string, error)                                                         //perm:read
	SendProposal(ctx context.Context, key string) error                                                         // perm:write
	StoreProposal(ctx context.Context, proposal types.OrderStoreProposal) (types2.StoreProposalResponse, error) // perm:write
	ShowProposal(ctx context.Context, cid string) (infos []types2.ProposalInfo, err error)                      // perm:read
}
