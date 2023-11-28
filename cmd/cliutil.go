package cliutil

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/SaoNetwork/sao-node/chain"
	"github.com/SaoNetwork/sao-node/node"
	"github.com/SaoNetwork/sao-node/types"
	"golang.org/x/term"
	"payment-gateway/api"
	apiclient "payment-gateway/api/client"
	"payment-gateway/payment-gateway/config"
	"payment-gateway/payment-gateway/repo"

	saodid "github.com/SaoNetwork/sao-did"
	saokey "github.com/SaoNetwork/sao-did/key"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/urfave/cli/v2"
)

const FlagKeyName = "key-name"
const APP_NAME_NODE = "saonode"
const APP_NAME_CLIENT = "saoclient"

var ApiToken string
var FlagToken = &cli.StringFlag{
	Name:        "token",
	Usage:       "connection token",
	EnvVars:     []string{"SAO_API_TOKEN"},
	Required:    false,
	Destination: &ApiToken,
}

var KeyringHome string
var FlagKeyringHome = &cli.StringFlag{
	Name:        "keyring",
	Usage:       "account keyring home directory",
	EnvVars:     []string{"SAO_KEYRING_HOME"},
	Value:       "~/.sao/",
	Destination: &KeyringHome,
}

var ChainAddress string
var FlagChainAddress = &cli.StringFlag{
	Name:        "chain-address",
	Usage:       "sao chain api",
	EnvVars:     []string{"SAO_CHAIN_API"},
	Destination: &ChainAddress,
}

// IsVeryVerbose is a global var signalling if the CLI is running in very
// verbose mode or not (default: false).
var IsVeryVerbose bool

// FlagVeryVerbose enables very verbose mode, which is useful when debugging
// the CLI itself. It should be included as a flag on the top-level command
// (e.g. saonode -vv).
var FlagVeryVerbose = &cli.BoolFlag{
	Name:        "vv",
	Usage:       "enables very verbose mode, useful for debugging the CLI",
	Destination: &IsVeryVerbose,
}

func AskForPassphrase() (string, error) {
	fmt.Print("Enter passphrase:")
	passphrase, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", types.Wrap(types.ErrInvalidPassphrase, err)
	}
	return string(passphrase), nil
}

func GetDidManager(cctx *cli.Context, address string) (*saodid.DidManager, error) {
	// repo := cctx.String("repo")
	//
	//address, err := chain.GetAddress(cctx.Context, KeyringHome, keyName)
	//if err != nil {
	//	return nil, "", err
	//}

	payload := fmt.Sprintf("cosmos %s allows to generate did", address)
	fmt.Println(payload)
	secret, err := chain.SignByAddress(cctx.Context, KeyringHome, address, []byte(payload))
	if err != nil {
		return nil, types.Wrap(types.ErrSignedFailed, err)
	}

	provider, err := saokey.NewSecp256k1Provider(secret)
	if err != nil {
		return nil, types.Wrap(types.ErrCreateProviderFailed, err)
	}
	resolver := saokey.NewKeyResolver()

	didManager := saodid.NewDidManager(provider, resolver)
	_, err = didManager.Authenticate([]string{}, "")
	if err != nil {
		return nil, types.Wrap(types.ErrAuthenticateFailed, err)
	}

	return &didManager, nil
}

func GetNodeApi(cctx *cli.Context, repoPath string, nodeApi string, apiToken string) (api.SaoApi, jsonrpc.ClientCloser, error) {
	if nodeApi != "" && apiToken != "" {
		gatewayApi, closer, err := apiclient.NewNodeApi(cctx.Context, nodeApi, apiToken)
		if err == nil {
			return gatewayApi, closer, err
		}
	}

	repo, err := repo.PrepareRepo(repoPath)
	if err != nil {
		return nil, nil, err
	}

	c, err := repo.Config()
	if err != nil {
		return nil, nil, types.Wrapf(types.ErrReadConfigFailed, "invalid config for repo, got: %T", c)
	}

	cfg, ok := c.(*config.Node)
	if !ok {
		return nil, nil, types.Wrapf(types.ErrDecodeConfigFailed, "invalid config for repo, got: %T", c)
	}

	key, err := repo.GetKeyBytes()
	if err != nil {
		return nil, nil, err
	}

	token, err := jwt.Sign(&node.JwtPayload{Allow: api.AllPermissions[:3]}, jwt.NewHS256(key))
	if err != nil {
		return nil, nil, types.Wrap(types.ErrSignedFailed, err)
	}

	ma, err := multiaddr.NewMultiaddr(cfg.Api.ListenAddress)
	if err != nil {
		return nil, nil, types.Wrap(types.ErrInvalidServerAddress, err)
	}
	_, addr, err := manet.DialArgs(ma)
	if err != nil {
		return nil, nil, types.Wrap(types.ErrConnectFailed, err)
	}

	apiAddress := "http://" + strings.ReplaceAll(addr, "0.0.0.0", "127.0.0.1") + "/rpc/v0"
	return apiclient.NewNodeApi(cctx.Context, apiAddress, string(token))
}

func GetChainAddress(cctx *cli.Context, repoPath string, binaryName string) (string, error) {
	if cctx.String("chain-address") != "" {
		return cctx.String("chain-address"), nil
	}

	chainAddress := ChainAddress

	if chainAddress == "" {
		if binaryName == "saopayment" {
			r, err := repo.PrepareRepo(repoPath)
			if err != nil {
				return chainAddress, types.Wrap(types.ErrInvalidRepoPath, err)
			}

			c, err := r.Config()
			if err != nil {
				return chainAddress, types.Wrap(types.ErrReadConfigFailed, err)
			}

			cfg, ok := c.(*config.Node)
			if !ok {
				return chainAddress, types.Wrap(types.ErrDecodeConfigFailed, err)
			}

			chainAddress = cfg.Chain.Remote
		} else {
			return chainAddress, types.Wrapf(types.ErrInvalidParameters, "invalid binary name %s", binaryName)
		}
	}

	fmt.Println("chainaddress: ", chainAddress)

	if chainAddress == "" {
		return chainAddress, types.Wrapf(types.ErrInvalidChainAddress, "no chain address specified")
	}

	return chainAddress, nil
}
