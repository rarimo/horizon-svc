package evm

import (
	"github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/config"
)

type evmChainGateway struct {
	cfg config.ChainGatewayCfg
}

func New(cfg config.ChainGatewayCfg) types.ChainGateway {
	return &evmChainGateway{
		cfg: cfg,
	}
}
