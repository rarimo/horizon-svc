package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/utils"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	gobind "gitlab.com/rarimo/contracts/evm-bridge/gobind/contracts/interfaces/handlers"
)

type nativeHandler struct {
	log       *logan.Entry
	cli       *ethclient.Client
	handler   *gobind.INativeHandler
	cursorer  types.Cursorer
	publisher services.QPublisher
}

func newNativeHandler(log *logan.Entry, cli *ethclient.Client, kv *redis.KeyValueProvider, publisher services.QPublisher, contractAddress common.Address, cursorKey, initialCursor string) Handler {
	handler, err := gobind.NewINativeHandler(contractAddress, cli)
	if err != nil {
		panic(errors.Wrap(err, "failed to init handler", logan.F{
			"handler": HandlerNative,
		}))
	}

	return &nativeHandler{
		log.WithField("handler", HandlerNative),
		cli,
		handler,
		utils.NewCursorer(log, kv, cursorKey+"_"+HandlerNative, initialCursor),
		publisher,
	}
}

func (h *nativeHandler) Name() string {
	return HandlerNative
}

func (h *nativeHandler) Run(ctx context.Context) error {
	start, err := h.cursorer.GetStartCursor(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get start cursor")
	}

	iter, err := h.handler.FilterWithdrawnNative(&bind.FilterOpts{
		Start:   uint64(start),
		Context: ctx,
	})
	if err != nil {
		return errors.Wrap(err, "failed to init native iterator")
	}

	for iter.Next() {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}
		e := iter.Event

		if e == nil {
			h.log.Error("got nil event")
			continue
		}
		h.log.WithFields(logan.F{
			"tx_hash":   e.Raw.TxHash,
			"tx_index":  e.Raw.TxIndex,
			"log_index": e.Raw.Index,
		}).Debug("got event")

		msg, err := logToWithdrawal(ctx, h.cli, e.Raw)
		if err != nil {
			return errors.Wrap(err, "failed to parse log")
		}

		err = h.publisher.PublishMsgs(ctx, msg.Message())
		if err != nil {
			return errors.Wrap(err, "failed to publish message")
		}

		if err = h.cursorer.SetStartCursor(ctx, int64(e.Raw.BlockNumber)); err != nil {
			return errors.Wrap(err, "failed to set cursor")
		}
	}

	return nil
}
