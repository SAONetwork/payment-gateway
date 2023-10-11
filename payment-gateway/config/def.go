package config

func DefaultSaoNode() *Config {
	return &Config{
		Chain: Chain{
			Remote:     "http://localhost:26657",
			WsEndpoint: "/websocket",
		},
	}
}

func defCommon() Common {
	return Common{
		Chain: Chain{
			Remote:     "http://localhost:26657",
			WsEndpoint: "/websocket",
			TxPoolSize: 0,
		},
		Libp2p: Libp2p{
			ListenAddress: []string{
				"/ip4/0.0.0.0/tcp/5153",
			},
			AnnounceAddresses: []string{},
			PublicAddress:     "",
			IntranetIpEnable:  true,
			ExternalIpEnable:  true,
		},
		Transport: Transport{
			TransportListenAddress: []string{
				"/ip4/0.0.0.0/udp/5154",
			},
			StagingSapceSize: 32 * 1024 * 1024 * 1024,
		},
		Module: Module{
			GatewayEnable: false,
			StorageEnable: true,
			IndexerEnable: false,
		},
	}
}
