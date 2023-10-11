package api

import (
	"context"
	"time"

	"github.com/rarimo/horizon-svc/pkg/txbuild"

	"github.com/rarimo/horizon-svc/internal/data/cachedpg"

	"github.com/go-chi/chi"
	"github.com/rarimo/horizon-svc/internal/chain_gateway"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/proxy"
	"github.com/rarimo/horizon-svc/internal/services/api/handlers"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/logan/v3"
)

func Run(ctx context.Context, cfg config.Config) {
	r := chi.NewRouter()

	const slowRequestDurationThreshold = time.Second
	ape.DefaultMiddlewares(r, cfg.Log(), slowRequestDurationThreshold)

	storage := cfg.NewStorage()

	r.Use(
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
			handlers.CtxChainsQ(cfg.ChainsQ()),
			handlers.CtxStorage(storage),
			handlers.CtxCachedStorage(cachedpg.NewStorage(cfg.Log(), storage, cfg.RedisClient())),
			handlers.CtxBuilder(txbuild.NewMultiBuilder(cfg)),
			handlers.CtxCore(cfg.Core()),
			handlers.CtxProxyRepo(
				proxy.New(
					cfg.ChainsQ(),
					cfg.MetadataFetcher(),
					chain_gateway.New(cfg.ChainGateway())),
			),
		),
	)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/chains", func(r chi.Router) {
			r.Get("/", handlers.ChainList)
		})

		r.Route("/tokens", func(r chi.Router) { // TODO v2: aggregations+filters
			r.Route("/{token_index}", func(r chi.Router) {
				r.Route("/chains", func(r chi.Router) {
					r.Route("/{chain}", func(r chi.Router) {
						r.Get("/balance/{account_address}", handlers.Balance)
						r.Get("/nfts/{token_id}/metadata", handlers.NftMetadata)
						r.Get("/accounts/{account_address}/nfts", handlers.NftList)
					})
				})
			})
		})

		r.Route("/transfers", func(r chi.Router) {
			r.Get("/", handlers.TransferList)
			r.Get("/{id}", handlers.TransferByID)
			r.Get("/{hash}/withdrawal/sse", handlers.TransferWithdrawalByHash)
		})
		r.Post("/buildtx", handlers.BuildTx)
	})

	cfg.Log().WithFields(logan.F{
		"service": "api",
		"addr":    cfg.Listener().Addr(),
	}).Info("starting api")

	ape.Serve(ctx, r, cfg, ape.ServeOpts{})
}
