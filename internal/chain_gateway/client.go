package chain_gateway

import (
	"github.com/rarimo/horizon-svc/internal/chain_gateway/evm"
	"github.com/rarimo/horizon-svc/internal/chain_gateway/solana"
	gateway "github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/config"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type ChainGateway interface {
	Get(chain string) gateway.ChainGateway
}

func New(cfg *config.ChainGateway) ChainGateway {
	p := chainGateway{
		gateways: make(map[string]gateway.ChainGateway),
	}

	for _, gatewayCfg := range cfg.Gateways {
		switch gatewayCfg.Type {
		case tokenmanager.NetworkType_EVM:
			p.gateways[gatewayCfg.Name] = evm.New(gatewayCfg)
		case tokenmanager.NetworkType_Solana:
			p.gateways[gatewayCfg.Name] = solana.New(gatewayCfg)
		case tokenmanager.NetworkType_Near, tokenmanager.NetworkType_Other:
			// bup
			//  currently we don't support near and other chain types for gateways
		default:
			panic("Unsupported chain type")
		}
	}

	return &p
}

type chainGateway struct {
	gateways map[string]gateway.ChainGateway
}

func (r *chainGateway) Get(chain string) gateway.ChainGateway {
	return r.gateways[chain]
}
