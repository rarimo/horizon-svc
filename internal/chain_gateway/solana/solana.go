package solana

import (
	"github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/config"
)

type solanaChainGateway struct {
	cfg config.ChainGatewayCfg
}

func New(cfg config.ChainGatewayCfg) types.ChainGateway {
	return &solanaChainGateway{
		cfg: cfg,
	}
}
