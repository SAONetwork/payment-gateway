// Code generated by github.com/SaoNetwork/sao-node/gen/api. DO NOT EDIT.

package api

import (
	"context"
	types2 "payment-gateway/types"

	"github.com/SaoNetwork/sao-node/types"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"golang.org/x/xerrors"
)

var ErrNotSupported = xerrors.New("method not supported")

type SaoApiStruct struct {
	Internal struct {
		AuthNew func(p0 context.Context, p1 []auth.Permission) ([]byte, error) `perm:"admin"`

		AuthVerify func(p0 context.Context, p1 string) ([]auth.Permission, error) `perm:"none"`

		GetNodeAddress func(p0 context.Context) (string, error) `perm:"read"`

		SendProposal func(p0 context.Context, p1 string) error `perm:"write"`

		ShowProposal func(p0 context.Context, p1 string) ([]types2.ProposalInfo, error) `perm:"read"`

		StoreProposal func(p0 context.Context, p1 types.OrderStoreProposal) (types2.StoreProposalResponse, error) `perm:"write"`
	}
}

type SaoApiStub struct {
}

func (s *SaoApiStruct) AuthNew(p0 context.Context, p1 []auth.Permission) ([]byte, error) {
	if s.Internal.AuthNew == nil {
		return *new([]byte), ErrNotSupported
	}
	return s.Internal.AuthNew(p0, p1)
}

func (s *SaoApiStub) AuthNew(p0 context.Context, p1 []auth.Permission) ([]byte, error) {
	return *new([]byte), ErrNotSupported
}

func (s *SaoApiStruct) AuthVerify(p0 context.Context, p1 string) ([]auth.Permission, error) {
	if s.Internal.AuthVerify == nil {
		return *new([]auth.Permission), ErrNotSupported
	}
	return s.Internal.AuthVerify(p0, p1)
}

func (s *SaoApiStub) AuthVerify(p0 context.Context, p1 string) ([]auth.Permission, error) {
	return *new([]auth.Permission), ErrNotSupported
}

func (s *SaoApiStruct) GetNodeAddress(p0 context.Context) (string, error) {
	if s.Internal.GetNodeAddress == nil {
		return "", ErrNotSupported
	}
	return s.Internal.GetNodeAddress(p0)
}

func (s *SaoApiStub) GetNodeAddress(p0 context.Context) (string, error) {
	return "", ErrNotSupported
}

func (s *SaoApiStruct) SendProposal(p0 context.Context, p1 string) error {
	if s.Internal.SendProposal == nil {
		return ErrNotSupported
	}
	return s.Internal.SendProposal(p0, p1)
}

func (s *SaoApiStub) SendProposal(p0 context.Context, p1 string) error {
	return ErrNotSupported
}

func (s *SaoApiStruct) ShowProposal(p0 context.Context, p1 string) ([]types2.ProposalInfo, error) {
	if s.Internal.ShowProposal == nil {
		return *new([]types2.ProposalInfo), ErrNotSupported
	}
	return s.Internal.ShowProposal(p0, p1)
}

func (s *SaoApiStub) ShowProposal(p0 context.Context, p1 string) ([]types2.ProposalInfo, error) {
	return *new([]types2.ProposalInfo), ErrNotSupported
}

func (s *SaoApiStruct) StoreProposal(p0 context.Context, p1 types.OrderStoreProposal) (types2.StoreProposalResponse, error) {
	if s.Internal.StoreProposal == nil {
		return *new(types2.StoreProposalResponse), ErrNotSupported
	}
	return s.Internal.StoreProposal(p0, p1)
}

func (s *SaoApiStub) StoreProposal(p0 context.Context, p1 types.OrderStoreProposal) (types2.StoreProposalResponse, error) {
	return *new(types2.StoreProposalResponse), ErrNotSupported
}

var _ SaoApi = new(SaoApiStruct)
