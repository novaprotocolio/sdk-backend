package ethereum

import (
	"fmt"
	"os"
)

type EthereumNova struct {
	*Ethereum
	*EthereumNovaProtocol
}

func NewEthereumNova(rpcURL, hybridExAddr string) *EthereumNova {
	if rpcURL == "" {
		rpcURL = os.Getenv("NSK_BLOCKCHAIN_RPC_URL")
	}

	if rpcURL == "" {
		panic(fmt.Errorf("NewEthereumNova need argument rpcURL"))
	}

	return &EthereumNova{
		NewEthereum(rpcURL, hybridExAddr),
		&EthereumNovaProtocol{},
	}
}
