package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"strconv"
	"sync"
	"time"
)

const (
	HandlerNative  = "native"
	HandlerERC20   = "erc20"
	HandlerERC721  = "erc721"
	HandlerERC1155 = "erc1155"
)

type evmProducer struct {
	log       *logan.Entry
	cfg       *config.BridgeProducerChainConfig
	chain     data.Chain
	publisher services.QPublisher
	handlers  []Handler
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	chain data.Chain,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	cursorKey,
	contractAddress string,
) types.Producer {
	f := logan.F{
		"chain": chain.Name,
		"rpc":   chain.Rpc,
	}

	cli, err := ethclient.Dial(chain.Rpc)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to ethereum node", f))
	}

	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		lastBlockHeight, err := cli.BlockNumber(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "failed to get last block height", f))
		}

		initialCursor = strconv.FormatUint(lastBlockHeight, 10)
	}

	contract := common.HexToAddress(contractAddress)

	handlers := []Handler{
		newNativeHandler(log, cli, chain.Name, kv, publisher, contract, cursorKey, initialCursor),
		newERC20Handler(log, cli, chain.Name, kv, publisher, contract, cursorKey, initialCursor),
		newERC721Handler(log, cli, chain.Name, kv, publisher, contract, cursorKey, initialCursor),
		newERC1155Handler(log, cli, chain.Name, kv, publisher, contract, cursorKey, initialCursor),
	}

	return &evmProducer{
		log.WithField("who", chain.Name+"_evm_bridge_events_producer"),
		cfg,
		chain,
		publisher,
		handlers,
	}
}

type Handler interface {
	Run(ctx context.Context) error
	Name() string
}

func (p *evmProducer) Run(ctx context.Context) error {
	wg := &sync.WaitGroup{}

	for _, handler := range p.handlers {
		wg.Add(1)

		p.log.WithField("handler", handler.Name()).Info("starting handler")

		go func(handler Handler) {
			defer wg.Done()
			running.WithBackOff(ctx, p.log, handler.Name()+"_handler", handler.Run, 5*time.Second, 10*time.Second, time.Minute)
		}(handler)
	}
	wg.Wait()

	return nil
}
