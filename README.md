### Installation and Setup
#### Install Go 1.19.1

Currently, SAO Network uses Go 1.19.1 to compile the code.

Install [Go 1.19.1](https://go.dev/doc/install) by following instructions there.

Verify the installation by typing `go version` in your terminal.

```
$ go version
go version go1.19.1 darwin/amd64
```



#### Build SAO Payment gateway 

```bash
$ git clone https://github.com/SaoNetwork/payment-gateway.git
$ make 
```

#### Faucet

In order to get testnet tokens use [https://faucet.testnet.sao.network/](https://faucet.testnet.sao.network/)

#### Join Network 

initialize your payment gateway 

```
$ ./saopayment --chain-address <SAO Chain RPC URL> init --creator <SAO Wallet> --payee <PAYEE CONTRACT> --provider <EVM Chain RPC URL> --height <EVENT FILTER FROm>
```


run node

```
$ ./saopayment run 
```



## License

Copyright © SAO Network, Inc. All rights reserved.

Licensed under the [Apache v2 License](LICENSE.md).
