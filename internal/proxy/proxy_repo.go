package proxy

import (
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/evm"
	"github.com/rarimo/horizon-svc/internal/proxy/near"
	"github.com/rarimo/horizon-svc/internal/proxy/solana"
	proxy "github.com/rarimo/horizon-svc/internal/proxy/types"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type ProxyRepo interface {
	Get(networkType string) proxy.Proxy
}

func New(chains data.ChainsQ, fetcher metadata_fetcher.Client) ProxyRepo {
	repo := proxyRepo{
		proxies: make(map[string]proxy.Proxy),
	}

	for _, chain := range chains.List() {
		switch chain.Type {
		case tokenmanager.NetworkType_EVM:
			repo.proxies[chain.Name] = evm.New(chain, fetcher)
		case tokenmanager.NetworkType_Solana:
			repo.proxies[chain.Name] = solana.New(chain, fetcher)
		case tokenmanager.NetworkType_Near:
			repo.proxies[chain.Name] = near.New(chain, fetcher)
		case tokenmanager.NetworkType_Other:
			// bup
			//  currently we don't support other chain types for proxying
		default:
			panic("Unsupported chain type")
		}
	}

	return &repo
}

type proxyRepo struct {
	proxies map[string]proxy.Proxy
}

func (r *proxyRepo) Get(chain string) proxy.Proxy {
	return r.proxies[chain]
}
