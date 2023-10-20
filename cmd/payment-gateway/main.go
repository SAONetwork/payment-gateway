package main

// TODO:
// * init should store node address locally.
// later cmd(join, quit) should call node process api to get node address if accountAddress not provided.

import (
	"bufio"
	"fmt"
	payment_gateway "payment-gateway/payment-gateway"
	"strings"

	"github.com/SaoNetwork/sao-node/build"
	cliutil "github.com/SaoNetwork/sao-node/cmd"
	"github.com/SaoNetwork/sao-node/cmd/account"
	"github.com/SaoNetwork/sao-node/node"
	"github.com/SaoNetwork/sao-node/node/repo"
	"github.com/SaoNetwork/sao-node/types"

	"cosmossdk.io/math"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"

	"github.com/ipfs/go-datastore"

	"os"

	"github.com/SaoNetwork/sao-node/chain"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

var log = logging.Logger("node")

const (
	FlagStorageRepo        = "repo"
	FlagStorageDefaultRepo = "~/.sao-node"
)

var NodeApi string
var FlagNodeApi = &cli.StringFlag{
	Name:        "node",
	Usage:       "node connection",
	EnvVars:     []string{"SAO_NODE_API"},
	Required:    false,
	Destination: &NodeApi,
}

var FlagRepo = &cli.StringFlag{
	Name:    FlagStorageRepo,
	Usage:   "repo directory for sao storage node",
	EnvVars: []string{"SAO_NODE_PATH"},
	Value:   FlagStorageDefaultRepo,
}

func before(_ *cli.Context) error {
	_ = logging.SetLogLevel("cache", "INFO")
	_ = logging.SetLogLevel("model", "INFO")
	_ = logging.SetLogLevel("node", "INFO")
	_ = logging.SetLogLevel("rpc", "INFO")
	_ = logging.SetLogLevel("chain", "INFO")
	_ = logging.SetLogLevel("gateway", "INFO")
	_ = logging.SetLogLevel("storage", "INFO")
	_ = logging.SetLogLevel("transport", "INFO")
	_ = logging.SetLogLevel("store", "INFO")
	_ = logging.SetLogLevel("indexer", "INFO")
	_ = logging.SetLogLevel("graphql", "INFO")
	if cliutil.IsVeryVerbose {
		_ = logging.SetLogLevel("cache", "DEBUG")
		_ = logging.SetLogLevel("model", "DEBUG")
		_ = logging.SetLogLevel("node", "DEBUG")
		_ = logging.SetLogLevel("rpc", "DEBUG")
		_ = logging.SetLogLevel("chain", "DEBUG")
		_ = logging.SetLogLevel("gateway", "DEBUG")
		_ = logging.SetLogLevel("storage", "DEBUG")
		_ = logging.SetLogLevel("transport", "DEBUG")
		_ = logging.SetLogLevel("store", "DEBUG")
		_ = logging.SetLogLevel("indexer", "DEBUG")
		_ = logging.SetLogLevel("graphql", "DEBUG")
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:                 cliutil.APP_NAME_NODE,
		Usage:                "Command line for sao network node",
		EnableBashCompletion: true,
		Version:              build.UserVersion(),
		Before:               before,
		Flags: []cli.Flag{
			FlagRepo,
			cliutil.FlagChainAddress,
			cliutil.FlagVeryVerbose,
			cliutil.FlagKeyringHome,
			FlagNodeApi,
			cliutil.FlagToken,
		},
		Commands: []*cli.Command{
			initCmd,
			joinCmd,
			cleanCmd,
			runCmd,
			infoCmd,
			account.AccountCmd,
			cliutil.GenerateDocCmd,
		},
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "initialize a sao network node",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "creator",
			Usage:    "node's account on sao chain",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multiaddr",
			Usage:    "nodes' multiaddr",
			Value:    "/ip4/127.0.0.1/tcp/5153/",
			Required: false,
		},
		&cli.UintFlag{
			Name:     "tx-pool-size",
			Usage:    "address pool size for sending message, the default value is 0",
			Value:    0,
			Required: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		chainAddress := cliutil.ChainAddress
		if chainAddress == "" {
			return types.Wrapf(types.ErrInvalidParameters, "must provide --chain-address")
		}

		repoPath := cctx.String(FlagStorageRepo)
		creator := cctx.String("creator")
		txPoolSize := cctx.Uint("tx-pool-size")

		r, err := initRepo(repoPath, chainAddress, txPoolSize)
		if err != nil {
			return err
		}

		c, err := r.Config()
		if err != nil {
			return types.Wrapf(types.ErrReadConfigFailed, "invalid config for repo, got: %T", c)
		}

		// init metadata datastore
		mds, err := r.Datastore(ctx, "/metadata")
		if err != nil {
			return types.Wrap(types.ErrOpenDataStoreFailed, err)
		}
		if err := mds.Put(ctx, datastore.NewKey("node-address"), []byte(creator)); err != nil {
			return types.Wrap(types.ErrGetFailed, err)
		}

		log.Info("initialize libp2p identity")

		chainSvc, err := chain.NewChainSvc(ctx, chainAddress, "/websocket", cliutil.KeyringHome)
		if err != nil {
			return err
		}

		for {
			fmt.Printf("Please make sure there is enough SAO tokens in the account %s. Confirm with 'yes' :", creator)

			reader := bufio.NewReader(os.Stdin)
			indata, err := reader.ReadBytes('\n')
			if err != nil {
				return types.Wrap(types.ErrInvalidParameters, err)
			}
			if strings.ToLower(strings.Replace(string(indata), "\n", "", -1)) != "yes" {
				continue
			}

			coins, err := chainSvc.GetBalance(ctx, creator)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			} else {
				if coins.AmountOf(chain.DENOM).LT(math.NewInt(int64(1100))) {
					continue
				} else {
					break
				}
			}

		}

		if tx, err := chainSvc.Create(ctx, creator); err != nil {
			// TODO: clear dir
			return err
		} else {
			fmt.Println(tx)
		}

		if txPoolSize > 0 {
			err = chain.CreateAddressPool(ctx, cliutil.KeyringHome, txPoolSize)
			if err != nil {
				return err
			}

			ap, err := chain.LoadAddressPool(ctx, cliutil.KeyringHome, txPoolSize)
			if err != nil {
				return err
			}

			for address := range ap.Addresses {
				amount := int64(1000 / txPoolSize)
				if tx, err := chainSvc.Send(ctx, creator, address, amount); err != nil {
					// TODO: clear dir
					return err
				} else {
					fmt.Printf("Sent %d SAO from creator %s to pool address %s, txhash=%s\r", amount, creator, address, tx)
				}
			}
		}

		return nil
	},
}

func initRepo(repoPath string, chainAddress string, TxPoolSize uint) (*repo.Repo, error) {
	// init base dir
	r, err := repo.NewRepo(repoPath)
	if err != nil {
		return nil, err
	}

	ok, err := r.Exists()
	if err != nil {
		return nil, types.Wrap(types.ErrOpenRepoFailed, err)
	}

	if ok {
		return nil, types.Wrapf(types.ErrInitRepoFailed, "repo at '%s' is already initialized", repoPath)
	}

	log.Info("Initializing repo")
	if err = r.Init(chainAddress, TxPoolSize); err != nil {
		return nil, err
	}
	return r, nil
}

var joinCmd = &cli.Command{
	Name:  "join",
	Usage: "join sao network",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "creator",
			Usage:    "node's account on sao chain",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		chainAddress, err := cliutil.GetChainAddress(cctx, cctx.String("repo"), cctx.App.Name)
		if err != nil {
			log.Warn(err)
		}
		creator := cctx.String("creator")

		chain, err := chain.NewChainSvc(ctx, chainAddress, "/websocket", cliutil.KeyringHome)
		if err != nil {
			return err
		}

		repo, err := prepareRepo(cctx)
		if err != nil {
			return err
		}
		c, err := repo.Config()
		if err != nil {
			return types.Wrapf(types.ErrReadConfigFailed, "invalid config for repo, got: %T", c)
		}

		// update metadata datastore
		mds, err := repo.Datastore(ctx, "/metadata")
		if err != nil {
			return types.Wrap(types.ErrOpenDataStoreFailed, err)
		}
		if err := mds.Put(ctx, datastore.NewKey("node-address"), []byte(creator)); err != nil {
			return types.Wrap(types.ErrGetFailed, err)
		}

		tx, err := chain.Create(ctx, creator)
		if err != nil {
			return err
		} else {
			fmt.Println(tx)
		}

		return nil
	},
}

var cleanCmd = &cli.Command{
	Name:  "clean",
	Usage: "clean up the local datastore",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		console := color.New(color.FgRed, color.Bold)
		console.Println("!!!BE CAREFULL!!!")
		console.Print("It'll remove all the configurations in the local datastore and you have to init a new storage node. Confirm with 'yes' :")
		reader := bufio.NewReader(os.Stdin)
		indata, err := reader.ReadBytes('\n')
		if err != nil {
			return types.Wrap(types.ErrInvalidParameters, err)
		}
		if strings.ToLower(strings.Replace(string(indata), "\n", "", -1)) == "yes" {
			repo, err := prepareRepo(cctx)
			if err != nil {
				return err
			}

			mds, err := repo.Datastore(ctx, "/metadata")
			if err != nil {
				return types.Wrap(types.ErrOpenDataStoreFailed, err)
			}
			mds.Delete(ctx, datastore.NewKey("node-address"))
			console.Println("Node address information has been deleted!")

			tds, err := repo.Datastore(ctx, "/transport")
			if err != nil {
				return types.Wrap(types.ErrOpenDataStoreFailed, err)
			}
			tds.Delete(ctx, datastore.NewKey(fmt.Sprintf(types.PEER_INFO_PREFIX)))
			console.Println("Peer information has been deleted!")
		}

		return nil
	},
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "start node",
	Action: func(cctx *cli.Context) error {
		myFigure := figure.NewFigure("Sao Network", "", true)
		myFigure.Print()

		// there is no place to trigger shutdown signal now. may add somewhere later.
		shutdownChan := make(chan struct{})
		ctx := cctx.Context

		repo, err := prepareRepo(cctx)
		if err != nil {
			return err
		}

		saopayment, err := payment_gateway.NewPaymentGateway(ctx, repo, cliutil.KeyringHome)
		if err != nil {
			return err
		}

		finishCh := node.MonitorShutdown(
			shutdownChan,
			node.ShutdownHandler{Component: "paymentgateway", StopFunc: saopayment.Stop},
		)
		<-finishCh
		return nil
	},
}

var infoCmd = &cli.Command{
	Name:  "info",
	Usage: "show node information",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "creator",
			Usage:    "node's account on sao chain",
			Required: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		chainAddress, err := cliutil.GetChainAddress(cctx, cctx.String("repo"), cctx.App.Name)
		if err != nil {
			log.Warn(err)
		}

		chain, err := chain.NewChainSvc(ctx, chainAddress, "/websocket", cliutil.KeyringHome)
		if err != nil {
			return err
		}

		creator := cctx.String("creator")
		if creator == "" {
			apiClient, closer, err := cliutil.GetNodeApi(cctx, cctx.String(FlagStorageRepo), NodeApi, cliutil.ApiToken)
			if err != nil {
				return types.Wrap(types.ErrCreateClientFailed, err)
			}
			defer closer()

			creator, err = apiClient.GetNodeAddress(ctx)
			if err != nil {
				return err
			}
		}
		chain.ShowBalance(ctx, creator)
		chain.ShowNodeInfo(ctx, creator)

		return nil
	},
}

func prepareRepo(cctx *cli.Context) (*repo.Repo, error) {
	return repo.PrepareRepo(cctx.String(FlagStorageRepo))
}