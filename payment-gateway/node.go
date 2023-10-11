package payment_gateway

import (
	"context"
	"fmt"
	saodid "github.com/SaoNetwork/sao-did"
	saokey "github.com/SaoNetwork/sao-did/key"
	"github.com/SaoNetwork/sao-did/sid"
	saodidtypes "github.com/SaoNetwork/sao-did/types"
	"github.com/SaoNetwork/sao-node/api"
	"github.com/SaoNetwork/sao-node/chain"
	"github.com/SaoNetwork/sao-node/node/config"
	"github.com/SaoNetwork/sao-node/node/repo"
	"github.com/SaoNetwork/sao-node/types"
	saotypes "github.com/SaoNetwork/sao/x/sao/types"
	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gbrlsnchs/jwt/v3"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("node")

const NODE_STATUS_NA uint32 = 0
const NODE_STATUS_ONLINE uint32 = 1
const NODE_STATUS_SERVE_GATEWAY uint32 = 1 << 1
const NODE_STATUS_SERVE_STORAGE uint32 = 1 << 2
const NODE_STATUS_ACCEPT_ORDER uint32 = 1 << 3
const NODE_STATUS_SERVE_INDEXER uint32 = 1 << 4
const NODE_STATUS_SERVE_PAYMENT uint32 = 1 << 5

type PaymentGateway struct {
	ctx       context.Context
	cfg       *config.Node
	repo      *repo.Repo
	address   string
	stopFuncs []StopFunc
	// used by store module
	chainSvc *chain.ChainSvc
}

type JwtPayload struct {
	Allow []auth.Permission
}

func NewPaymentGateway(ctx context.Context, repo *repo.Repo, keyringHome string) (*PaymentGateway, error) {
	c, err := repo.Config()
	if err != nil {
		return nil, err
	}

	cfg, ok := c.(*config.Node)
	if !ok {
		return nil, types.Wrapf(types.ErrDecodeConfigFailed, "invalid config for repo, got: %T", c)
	}

	// get node address
	mds, err := repo.Datastore(ctx, "/metadata")
	if err != nil {
		return nil, err
	}
	abytes, err := mds.Get(ctx, datastore.NewKey("node-address"))
	if err != nil {
		return nil, types.Wrap(types.ErrGetFailed, err)
	}
	nodeAddr := string(abytes)

	// chain
	chainSvc, err := chain.NewChainSvc(ctx, cfg.Chain.Remote, cfg.Chain.WsEndpoint, keyringHome)
	if err != nil {
		return nil, err
	}

	var stopFuncs []StopFunc
	sn := PaymentGateway{
		ctx:       ctx,
		cfg:       cfg,
		repo:      repo,
		address:   nodeAddr,
		stopFuncs: stopFuncs,
		chainSvc:  chainSvc,
	}

	var status = NODE_STATUS_ONLINE | NODE_STATUS_SERVE_PAYMENT
	notifyChan := make(map[string]chan interface{})

	log.Info("store manager daemon initialized")

	tokenRead, err := sn.AuthNew(ctx, api.AllPermissions[:2])
	if err != nil {
		return nil, err
	}
	log.Info("Read token: ", string(tokenRead))

	tokenWrite, err := sn.AuthNew(ctx, api.AllPermissions[:3])
	if err != nil {
		return nil, err
	}
	log.Info("Write token: ", string(tokenWrite))

	// chainSvc.stop should be after chain listener unsubscribe
	sn.stopFuncs = append(sn.stopFuncs, chainSvc.Stop)

	chainSvc.Reset(ctx, sn.address, "", status, nil, nil)
	log.Infof("repo: %s, Remote: %s, WsEndpoint： %s", repo.Path, cfg.Chain.Remote, cfg.Chain.WsEndpoint)
	log.Infof("node[%s] is joining SAO network...", sn.address)
	if err != nil {
		return nil, err
	}

	chainSvc.StartStatusReporter(ctx, sn.address, status)

	sn.stopFuncs = append(sn.stopFuncs, func(_ context.Context) error {
		for _, c := range notifyChan {
			close(c)
		}
		return nil
	})

	return &sn, nil
}

func (n *PaymentGateway) Stop(ctx context.Context) error {
	for _, f := range n.stopFuncs {
		err := f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *PaymentGateway) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	var payload JwtPayload
	key, err := n.repo.GetKeyBytes()
	if err != nil {
		return nil, types.Wrap(types.ErrDecodeConfigFailed, err)
	}

	if _, err := jwt.Verify([]byte(token), jwt.NewHS256(key), &payload); err != nil {
		return nil, types.Wrapf(types.ErrInvalidJwt, "JWT Verification failed: %v", err)
	}

	log.Info("Permissions: ", payload)

	return payload.Allow, nil
}

func (n *PaymentGateway) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	p := JwtPayload{
		Allow: perms, // TODO: consider checking validity
	}

	key, err := n.repo.GetKeyBytes()
	if err != nil {
		return nil, types.Wrap(types.ErrDecodeConfigFailed, err)
	}
	return jwt.Sign(&p, jwt.NewHS256(key))
}

func (n *PaymentGateway) GetNodeAddress(ctx context.Context) (string, error) {
	return n.address, nil
}

func (n *PaymentGateway) getSidDocFunc() func(versionId string) (*sid.SidDocument, error) {
	return func(versionId string) (*sid.SidDocument, error) {
		return n.chainSvc.GetSidDocument(n.ctx, versionId)
	}
}

func (n *PaymentGateway) validSignature(ctx context.Context, proposal types.ConsensusProposal, owner string, signature saotypes.JwsSignature) error {
	if owner == "all" {
		return nil
	}

	didManager, err := saodid.NewDidManagerWithDid(owner, n.getSidDocFunc())
	if err != nil {
		return types.Wrap(types.ErrInvalidDid, err)
	}

	proposalBytes, err := proposal.Marshal()
	if err != nil {
		return types.Wrap(types.ErrMarshalFailed, err)
	}

	// log.Error("base64url.Encode(proposalBytes): ", base64url.Encode(proposalBytes))
	// log.Error("proposal: %#v", proposal)
	_, err = didManager.VerifyJWS(saodidtypes.GeneralJWS{
		Payload: base64url.Encode(proposalBytes),
		Signatures: []saodidtypes.JwsSignature{
			saodidtypes.JwsSignature(signature),
		},
	})
	if err != nil {
		return types.Wrap(types.ErrInvalidSignature, err)
	}

	return nil
}

func (n *PaymentGateway) getDidManager(ctx context.Context, keyringHome string, keyName string) (*saodid.DidManager, string, error) {
	fmt.Println("keyName: ", keyName)

	address, err := chain.GetAddress(ctx, keyringHome, keyName)
	if err != nil {
		return nil, "", err
	}

	payload := fmt.Sprintf("cosmos %s allows to generate did", address)
	secret, err := chain.SignByAccount(ctx, keyringHome, keyName, []byte(payload))
	if err != nil {
		return nil, "", types.Wrap(types.ErrSignedFailed, err)
	}

	provider, err := saokey.NewSecp256k1Provider(secret)
	if err != nil {
		return nil, "", types.Wrap(types.ErrCreateProviderFailed, err)
	}
	resolver := saokey.NewKeyResolver()

	didManager := saodid.NewDidManager(provider, resolver)
	_, err = didManager.Authenticate([]string{}, "")
	if err != nil {
		return nil, "", types.Wrap(types.ErrAuthenticateFailed, err)
	}

	return &didManager, address, nil
}

func (n *PaymentGateway) SendProposal(ctx context.Context, proposal saotypes.Proposal, signature saotypes.JwsSignature) error {
	// check meta?
	addr, err := n.chainSvc.QueryPaymentAddress(ctx, proposal.PaymentDid)
	if err != nil {
		return err
	}
	if addr != n.address {
		return types.ErrInvalidServerAddress
	}

	err = n.validSignature(ctx, &proposal, proposal.Owner, signature)
	if err != nil {
		return err
	}

	_, _, _, err = n.chainSvc.StoreOrder(ctx, n.address, &types.OrderStoreProposal{
		Proposal:     proposal,
		JwsSignature: signature,
	})
	return err
}
