package config

import (
	"github.com/rarimo/near-go/nearprovider"
	"net/http"

	"github.com/rarimo/horizon-svc/internal/data/cachedpg"

	"github.com/rarimo/horizon-svc/internal/core"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/mem"
	"github.com/rarimo/horizon-svc/internal/data/pg"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/pkg/ipfs"
	"github.com/rarimo/horizon-svc/pkg/rd"
	thttp "github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/copus"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"google.golang.org/grpc"
)

type Config interface {
	comfig.Logger
	pgdb.Databaser
	types.Copuser
	comfig.Listenerer
	ipfs.IPFSer
	rd.Rediser
	mem.Chainer
	nearprovider.Nearer

	NewStorage() data.Storage
	CachedStorage() data.Storage
	ChainGateway() *ChainGateway
	MetadataFetcher() metadata_fetcher.Client
	Core() core.Core
	Tendermint() *thttp.HTTP
	Cosmos() *grpc.ClientConn
	TransfersIndexer() TransfersIndexerConfig

	BlockRangeProducer() *BlockRangesProducerConfig

	RarimoCoreProducer() *RarimoCoreProducerConfig
	ConfirmationsIndexer() ConfirmationsIndexerConfig
	ApprovalsIndexer() ApprovalsIndexerConfig
	RejectionsIndexer() RejectionsIndexerConfig
	VotesIndexer() VotesIndexerConfig

	TokenManagerProducer() *TokenManagerProducerConfig
	ItemsIndexer() ItemsIndexerConfig
	CollectionsIndexer() CollectionsIndexerConfig

	Genesis() GenesisConfig

	BridgeProducer() *BridgeProducerConfig
	WithdrawalsIndexer() *WithdrawalsIndexerConfig

	RateLimiter() *RateLimiterConfig
}

type config struct {
	comfig.Logger
	pgdb.Databaser
	types.Copuser
	comfig.Listenerer
	ipfs.IPFSer
	rd.Rediser
	mem.Chainer
	nearprovider.Nearer

	chainGateway         comfig.Once
	chains               comfig.Once
	metadataFetcher      comfig.Once
	tendermint           comfig.Once
	cosmos               comfig.Once
	core                 comfig.Once
	transfersIndexer     comfig.Once
	blockRangeProducer   comfig.Once
	rarimocoreProducer   comfig.Once
	tokenmanagerProducer comfig.Once
	confirmationsIndexer comfig.Once
	approvalsIndexer     comfig.Once
	rejectionsIndexer    comfig.Once
	votesIndexer         comfig.Once
	itemsIndexer         comfig.Once
	collectionsIndexer   comfig.Once
	cachedstorage        comfig.Once
	genesis              comfig.Once
	bridgeProducer       comfig.Once
	withdrawalsIndexer   comfig.Once
	rateLimiter          comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	logger := comfig.NewLogger(getter, comfig.LoggerOpts{})
	return &config{
		getter:     getter,
		Logger:     logger,
		Databaser:  pgdb.NewDatabaser(getter),
		Copuser:    copus.NewCopuser(getter),
		Listenerer: comfig.NewListenerer(getter),
		Rediser:    rd.NewRediser(getter),
		IPFSer:     ipfs.NewIPFSer(getter),
		Chainer:    mem.NewChainer(getter),
		Nearer:     nearprovider.NewNearer(getter, logger.Log()),
	}
}

func (c *config) NewStorage() data.Storage {
	return pg.New(c.DB().Clone())
}

func (c *config) CachedStorage() data.Storage {
	return c.cachedstorage.Do(func() interface{} {
		return cachedpg.NewStorage(c.Log(), c.NewStorage(), c.RedisClient())
	}).(data.Storage)
}

func (c *config) Core() core.Core {
	return c.core.Do(func() interface{} {
		return core.NewCore(c.Cosmos())
	}).(core.Core)
}

func (c *config) MetadataFetcher() metadata_fetcher.Client {
	return c.metadataFetcher.Do(func() interface{} {
		return metadata_fetcher.New(http.DefaultClient, c.IPFS())
	}).(metadata_fetcher.Client)
}
