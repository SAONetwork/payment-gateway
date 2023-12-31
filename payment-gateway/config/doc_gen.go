// Code generated by github.com/SaoNetwork/sao-node/gen/cfgdoc. DO NOT EDIT.

package config

type DocField struct {
	Name    string
	Type    string
	Comment string
}

var Doc = map[string][]DocField{
	"API": []DocField{
		{
			Name: "ListenAddress",
			Type: "string",

			Comment: `Binding address for the Sao Node API`,
		},
		{
			Name: "Timeout",
			Type: "time.Duration",

			Comment: ``,
		},
		{
			Name: "EnablePermission",
			Type: "bool",

			Comment: ``,
		},
	},
	"Chain": []DocField{
		{
			Name: "Remote",
			Type: "string",

			Comment: `remote connection string`,
		},
		{
			Name: "WsEndpoint",
			Type: "string",

			Comment: `websocket endpoint`,
		},
	},
	"Node": []DocField{
		{
			Name: "Chain",
			Type: "Chain",

			Comment: ``,
		},
		{
			Name: "Api",
			Type: "API",

			Comment: ``,
		},
	},
}
