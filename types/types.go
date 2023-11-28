package types

import "github.com/SaoNetwork/sao-node/types"

type ProposalInfo struct {
	Key   string
	Value types.OrderStoreProposal
}

type StoreProposalResponse struct {
	DataId      string
	ProposalCid string
}
