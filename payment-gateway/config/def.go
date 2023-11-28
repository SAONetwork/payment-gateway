package config

import "time"

func DefaultSaoNode() *Node {
	return &Node{
		Chain: Chain{
			Remote:     "http://localhost:26657",
			WsEndpoint: "/websocket",
		},
		Api: API{
			ListenAddress:    "/ip4/127.0.0.1/tcp/5161/http",
			Timeout:          30 * time.Second,
			EnablePermission: false,
		},
	}
}
