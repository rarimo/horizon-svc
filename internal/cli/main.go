package cli

import (
	"context"
	"fmt"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rarimo/horizon-svc/internal/services/api"

	"github.com/rarimo/horizon-svc/internal/services"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/alecthomas/kingpin"
	"github.com/rarimo/horizon-svc/internal/config"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
)

func Run(args []string) {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	cfg := config.New(kv.MustFromEnv())
	log = cfg.Log()

	app := kingpin.New("horizon", "")

	runCmd := app.Command("run", "run command")
	apiCmd := runCmd.Command("api", "run API")

	// commands to run everything in the same instance
	tokenmanagerIndexerCmd := runCmd.Command("tokenmanager_indexing", "run routines to index tokenmanager events (collections, items)")
	rarimocoreIndexerCmd := runCmd.Command("rarimocore_indexing", "run routines to index rarimocore events (transfers, votes, confirmations etc.)")
	bridgeIndexerCmd := runCmd.Command("bridge_indexing", "run routines to index bridge events (withdrawals)")

	// block range producer can be run multi-purposely
	blockRangeProducerCmd := runCmd.Command("block_range_producer", "run block range producer")

	// routines for indexing rarimocore submodule
	rarimocoreOpProducerCmd := runCmd.Command("rarimocore_operations_producer", "run rarimocore operations producer (transfers, votes, confirmations etc.)")
	transfersIndexerCmd := runCmd.Command("transfers_indexer", "run transfers indexer")
	approvalIndexerCmd := runCmd.Command("approval_indexer", "run approval indexer")
	rejectionIndexerCmd := runCmd.Command("rejection_indexer", "run rejection indexer")
	confirmationsIndexerCmd := runCmd.Command("confirmations_indexer", "run confirmations indexer")
	votesIndexerCmd := runCmd.Command("votes_indexer", "run votes indexer")

	// routines for indexing tokenmanager submodule
	tokenmanagerOpProducerCmd := runCmd.Command("tokenmanager_operations_producer", "run tokenmanager operations producer (collections, items)")
	itemsIndexerCmd := runCmd.Command("items_indexer", "run items' events indexer")
	collectionsIndexerCmd := runCmd.Command("collections_indexer", "run collections' events indexer")

	// routines for indexing bridge contracts
	bridgeProducerCmd := runCmd.Command("bridge_events_producer", "run bridge events producer")
	withdrawalsIndexerCmd := runCmd.Command("withdrawals_indexer", "run withdrawals indexer")

	migrateCmd := app.Command("migrate", "migrate command")
	migrateUpCmd := migrateCmd.Command("up", "migrate db up")
	migrateDownCmd := migrateCmd.Command("down", "migrate db down")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		panic(errors.Wrap(err, "failed to parse args"))
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	run := func(f func(context.Context, config.Config)) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() {
				if rvr := recover(); rvr != nil {
					logan.New().WithRecover(rvr).Error("service panicked")
				}
			}()

			f(ctx, cfg)
		}()
	}

	switch cmd {
	case apiCmd.FullCommand():
		cfg.Log().Info("starting API")
		run(api.Run)
	case tokenmanagerIndexerCmd.FullCommand():
		cfg.Log().Info("starting tokenmanager indexers")
		run(services.RunBlockRangeProducer)
		run(services.RunTokenManagerEventsProducer)
		run(services.RunItemsIndexer)
		run(services.RunCollectionsIndexer)
	case rarimocoreIndexerCmd.FullCommand():
		cfg.Log().Info("starting rarimocore indexers")
		run(services.RunBlockRangeProducer)
		run(services.RunRarimoCoreOpProducer)
		run(services.RunTransfersIndexer)
		run(services.RunApprovalIndexer)
		run(services.RunRejectionIndexer)
		run(services.RunConfirmationsIndexer)
		run(services.RunVotesIndexer)
	case bridgeIndexerCmd.FullCommand():
		cfg.Log().Info("starting bridge indexers")
		run(bridge_producer.RunBridgeEventsProducer)
		run(services.RunWithdrawalsIndexer)
	case blockRangeProducerCmd.FullCommand():
		cfg.Log().Info("starting block range producer")
		run(services.RunBlockRangeProducer)
	case rarimocoreOpProducerCmd.FullCommand():
		cfg.Log().Info("starting rarimocore operations producer")
		run(services.RunRarimoCoreOpProducer)
	case tokenmanagerOpProducerCmd.FullCommand():
		cfg.Log().Info("starting tokenmanager operations producer")
		run(services.RunTokenManagerEventsProducer)
	case itemsIndexerCmd.FullCommand():
		cfg.Log().Info("starting items indexer")
		run(services.RunItemsIndexer)
	case collectionsIndexerCmd.FullCommand():
		cfg.Log().Info("starting collection indexer")
		run(services.RunCollectionsIndexer)
	case votesIndexerCmd.FullCommand():
		cfg.Log().Info("starting votes indexer")
		run(services.RunVotesIndexer)
	case transfersIndexerCmd.FullCommand():
		cfg.Log().Info("starting transfers indexer")
		run(services.RunTransfersIndexer)
	case approvalIndexerCmd.FullCommand():
		cfg.Log().Info("starting approval indexer")
		run(services.RunApprovalIndexer)
	case rejectionIndexerCmd.FullCommand():
		cfg.Log().Info("starting rejection indexer")
		run(services.RunRejectionIndexer)
	case confirmationsIndexerCmd.FullCommand():
		cfg.Log().Info("starting confirmations indexer")
		run(services.RunConfirmationsIndexer)
	case withdrawalsIndexerCmd.FullCommand():
		cfg.Log().Info("starting withdrawals indexer")
		run(services.RunWithdrawalsIndexer)
	case bridgeProducerCmd.FullCommand():
		cfg.Log().Info("starting bridge events producer")
		run(bridge_producer.RunBridgeEventsProducer)
	case migrateUpCmd.FullCommand():
		if err := MigrateUp(cfg); err != nil {
			panic(errors.Wrap(err, "failed to migrate up"))
		}
	case migrateDownCmd.FullCommand():
		if err := MigrateDown(cfg); err != nil {
			panic(errors.Wrap(err, "failed to migrate down"))
		}
	default:
		panic(fmt.Errorf("unknown command %s", cmd))
	}

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	wgch := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgch)
	}()

	select {
	case <-wgch:
		cfg.Log().Warn("all services stopped")
	case <-gracefulStop:
		cfg.Log().Info("received signal to stop")
		cancel()
		<-wgch
	}
}
