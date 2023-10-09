package bridge_producer

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/evm"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/near"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/rarimo"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/solana"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func newProducerer(cfg config.Config, publisher services.QPublisher) types.Producerer {
	repo := producerer{
		producers: make(map[string]types.Producer),
	}

	kv := redis.NewKeyValueProvider(cfg)

	networks, err := cfg.Core().TokenManager().GetParams(context.Background())
	if err != nil {
		panic(errors.Wrap(err, "failed to get networks"))
	}

	contracts := make(map[string]string)
	for _, net := range networks.Networks {
		bridgeParams := net.GetBridgeParams()
		if bridgeParams == nil {
			continue
		}

		contracts[net.Name] = bridgeParams.Contract
	}

	for _, chain := range cfg.ChainsQ().List() {
		cursorKey := chain.Name + "_" + cfg.BridgeProducer().CursorKey
		log := cfg.Log().WithField("who", chain.Name+"_bridge_events_producer")
		conf := cfg.BridgeProducer().ChainConfigByID(chain.ID)

		switch chain.Type {
		case tokenmanager.NetworkType_EVM:
			repo.producers[chain.Name] = evm.New(conf, log, chain, kv, publisher, cursorKey)
		case tokenmanager.NetworkType_Solana:
			repo.producers[chain.Name] = solana.New(conf, log, chain, kv, publisher, cursorKey)
		case tokenmanager.NetworkType_Near:
			repo.producers[chain.Name] = near.New(conf, log, chain, kv, publisher, cfg.Near(), cursorKey)
		case tokenmanager.NetworkType_Other: // FIXME: do we need another type for the rarimo chain?
			repo.producers[chain.Name] = rarimo.New(conf, log, chain, kv, publisher, cfg.Tendermint(), cursorKey)
		default:
			panic("Unsupported chain type")
		}
	}

	return &repo
}

type producerer struct {
	producers map[string]types.Producer
}

func (r *producerer) Get(chain string) types.Producer {
	return r.producers[chain]
}
