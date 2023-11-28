package config

import "time"

type Node struct {
	Chain Chain
	Api   API
}

// API contains configs for API endpoint
type API struct {

	// Binding address for the Sao Node API
	ListenAddress string

	Timeout time.Duration

	EnablePermission bool
}

// Chain contains configs for sao chain information
type Chain struct {

	// remote connection string
	Remote string

	// websocket endpoint
	WsEndpoint string
}
