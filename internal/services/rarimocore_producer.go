package services

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.com/distributed_lab/logan/v3"
	"strconv"
	"strings"
	"time"

	"github.com/tendermint/tendermint/rpc/jsonrpc/types"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"github.com/rarimo/xo/types/xo"
	"github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunRarimoCoreOpProducer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.RarimoCoreProducer().RunnerName+"_rarimocore_events_producer")

	transfersPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.RarimoCoreProducer().RunnerName+"_transfers_publisher",
		cfg.RarimoCoreProducer().TransfersQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create transfers publisher"))
	}

	confirmationsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.RarimoCoreProducer().RunnerName+"_confirmations_publisher",
		cfg.RarimoCoreProducer().ConfirmationsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create confirmations publisher"))
	}

	approvalsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.RarimoCoreProducer().RunnerName+"_approvals_publisher",
		cfg.RarimoCoreProducer().ApprovalsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create approvals publisher"))
	}

	rejectionsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.RarimoCoreProducer().RunnerName+"_rejections_publisher",
		cfg.RarimoCoreProducer().RejectionsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create rejections publisher"))
	}

	votesPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.RarimoCoreProducer().RunnerName+"_votes_publisher",
		cfg.RarimoCoreProducer().VotesQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create votes publisher"))
	}

	producer := &rarimoCoreOpProducer{
		log:                    log,
		thttp:                  cfg.Tendermint(),
		transfersPublisher:     transfersPublisher,
		confirmationsPublisher: confirmationsPublisher,
		approvalsPublisher:     approvalsPublisher,
		rejectionsPublisher:    rejectionsPublisher,
		votesPublisher:         votesPublisher,
		kv:                     redis.NewKeyValueProvider(cfg),
		txQ:                    cfg.CachedStorage().Clone().TransactionQ(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.RarimoCoreProducer().BlockRangeConsumer,
		producer,
	).Run(ctx)
}

type rarimoCoreOpProducer struct {
	log   *logan.Entry
	thttp *http.HTTP

	transfersPublisher     QPublisher
	confirmationsPublisher QPublisher
	approvalsPublisher     QPublisher
	rejectionsPublisher    QPublisher
	votesPublisher         QPublisher

	kv  data.KeyValueQ
	txQ data.TransactionQ
}

func (p *rarimoCoreOpProducer) Handle(ctx context.Context, msgs []msgs.Message) error {
	for _, msg := range msgs {
		brMsg := msg.MustBlockRangeMessage()
		p.log.WithFields(logan.F{
			"start": brMsg.Start,
			"end":   brMsg.End,
		}).Info("received block range message")

		if err := p.produceMsgs(ctx, brMsg); err != nil {
			return errors.Wrap(err, "failed to produce op msgs", logan.F{
				"start": brMsg.Start,
				"end":   brMsg.End,
			})
		}
	}

	return nil
}

type eventMsgs struct {
	newTransferEvents     []msgs.Message
	newVoteEvents         []msgs.Message
	newApprovalEvents     []msgs.Message
	newRejectionEvents    []msgs.Message
	newConfirmationEvents []msgs.Message
}

func (p *rarimoCoreOpProducer) produceMsgs(ctx context.Context, blockRange msgs.BlockRangeMessage) error {
	cursorKey := rarimocoreBlockrangeCursor(blockRange.Start, blockRange.End)

	blockRangeCursor, err := p.kv.Get(ctx, cursorKey)
	if err != nil {
		return errors.Wrap(err, "failed to get block range cursor", logan.F{
			"key": cursorKey,
		})
	}

	if blockRangeCursor == nil {
		now := time.Now().UTC()
		blockRangeCursor = &data.KeyValue{
			Key:       cursorKey,
			Value:     "1",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	page64, err := strconv.ParseInt(blockRangeCursor.Value, 10, 0)
	if err != nil {
		return errors.Wrap(err, "failed to parse block range cursor value", logan.F{
			"key":   cursorKey,
			"value": blockRangeCursor.Value,
		})
	}

	page := int(page64)
	for {
		p.log.WithFields(logan.F{
			"page": page,
			"from": blockRange.Start,
			"to":   blockRange.End,
		}).Debug("fetching page of transactions")
		txs, err := getTransactionsPage(ctx, p.thttp, blockRange.Start, blockRange.End, page)
		if err != nil {
			return errors.Wrap(err, "failed to get transactions page", logan.F{
				"start": blockRange.Start,
				"end":   blockRange.End,
				"page":  page,
			})
		}

		if len(txs) == 0 {
			if err := p.kv.Remove(ctx, cursorKey); err != nil {
				p.log.WithField("cursor", cursorKey).Warn("failed to remove cursor after finishing block")
			}

			p.log.WithFields(logan.F{
				"start": blockRange.Start,
				"end":   blockRange.End,
			}).Info("no txs left, finished processing block range")
			break
		}

		p.log.WithFields(logan.F{
			"page": page,
			"from": blockRange.Start,
			"to":   blockRange.End,
			"txs":  len(txs),
		}).Debug("got transactions, producing operations")

		eventsSet := p.extractEvents(txs)

		f := logan.F{
			"start": blockRange.Start,
			"end":   blockRange.End,
		}

		p.log.WithFields(f.Merge(logan.F{
			"transfers": len(eventsSet.newTransferEvents),
		})).Debug("publishing transfers")

		if err := p.transfersPublisher.PublishMsgs(ctx, eventsSet.newTransferEvents...); err != nil {
			return errors.Wrap(err, "failed to publish transfer messages", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"confirmations": len(eventsSet.newConfirmationEvents),
		})).Debug("publishing confirmations")

		if err := p.confirmationsPublisher.PublishMsgs(ctx, eventsSet.newConfirmationEvents...); err != nil {
			return errors.Wrap(err, "failed to publish confirmation messages", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"approvals": len(eventsSet.newApprovalEvents),
		})).Debug("publishing approvals")

		if err := p.approvalsPublisher.PublishMsgs(ctx, eventsSet.newApprovalEvents...); err != nil {
			return errors.Wrap(err, "failed to publish approval messages", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"rejections": len(eventsSet.newRejectionEvents),
		})).Debug("publishing rejections")

		if err := p.rejectionsPublisher.PublishMsgs(ctx, eventsSet.newRejectionEvents...); err != nil {
			return errors.Wrap(err, "failed to publish rejection messages", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"votes": len(eventsSet.newVoteEvents),
		})).Debug("publishing votes")

		if err := p.votesPublisher.PublishMsgs(ctx, eventsSet.newVoteEvents...); err != nil {
			return errors.Wrap(err, "failed to publish vote messages", f)
		}

		p.log.WithFields(f).Debug("published messages")

		page += 1
		err = p.kv.Upsert(ctx, data.KeyValue{
			Key:       blockRangeCursor.Key,
			Value:     strconv.Itoa(page),
			CreatedAt: blockRangeCursor.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to update block range cursor", logan.F{
				"key":   blockRangeCursor.Key,
				"value": page,
			})
		}

		if err := p.txQ.InsertBatchCtx(ctx, mustToData(txs)...); err != nil {
			return errors.Wrap(err, "failed to insert transactions", f)
		}
	}

	return nil
}

type txWithBlockInfo struct {
	tx          *coretypes.ResultTx
	blockHeight int64
	blockTime   time.Time
}

func getTransactionsPage(ctx context.Context, thttp *http.HTTP, start, end int64, pageNum int) ([]txWithBlockInfo, error) {
	q := fmt.Sprintf("tx.height>=%d AND tx.height<%d", start, end)
	if start == end {
		q = fmt.Sprintf("tx.height=%d", start)
	}

	lim := 100
	resp, err := thttp.TxSearch(ctx,
		q,
		true,
		&pageNum,
		&lim,
		"asc")

	if err != nil {
		if isOutOfRangeError(err) {
			return nil, nil
		}

		// TODO check for error out of range and return nil, true, nil
		return nil, errors.Wrap(err, "failed to get transactions", logan.F{
			"start": start,
			"end":   end,
		})
	}

	if resp.TotalCount == 0 {
		return nil, nil
	}

	blocks := make(map[int64]*coretypes.ResultBlock)
	for _, tx := range resp.Txs {
		if _, ok := blocks[tx.Height]; !ok {
			block, err := thttp.Block(ctx, &tx.Height)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get block by height", logan.F{
					"height": tx.Height,
				})
			}

			blocks[tx.Height] = block
		}
	}

	result := make([]txWithBlockInfo, 0, len(resp.Txs))
	for _, tx := range resp.Txs {
		block := blocks[tx.Height]
		result = append(result, txWithBlockInfo{
			tx:          tx,
			blockHeight: block.Block.Height,
			blockTime:   block.Block.Time,
		})
	}

	return result, nil
}

func (p *rarimoCoreOpProducer) extractEvents(txs []txWithBlockInfo) eventMsgs {
	eventsSet := eventMsgs{
		newTransferEvents:     make([]msgs.Message, 0, len(txs)),
		newVoteEvents:         make([]msgs.Message, 0, len(txs)),
		newApprovalEvents:     make([]msgs.Message, 0, len(txs)),
		newRejectionEvents:    make([]msgs.Message, 0, len(txs)),
		newConfirmationEvents: make([]msgs.Message, 0, len(txs)),
	}

	for _, txInfo := range txs {
		for _, event := range txInfo.tx.TxResult.Events {
			switch event.Type {
			case rarimotypes.EventTypeNewOperation:
				eventsSet.newTransferEvents = append(eventsSet.newTransferEvents,
					mustTransferOpMsg(txInfo.tx.Hash.String(), cosmostypes.Event(event)).Message())
			case rarimotypes.EventTypeNewConfirmation:
				eventsSet.newConfirmationEvents = append(eventsSet.newConfirmationEvents,
					mustConfirmationOpMsg(txInfo.tx.Hash.String(), cosmostypes.Event(event)).Message())
			case rarimotypes.EventTypeVoted:
				eventsSet.newVoteEvents = append(eventsSet.newVoteEvents,
					mustVoteOpMsg(txInfo.tx.Hash.String(), cosmostypes.Event(event)).Message())
			case rarimotypes.EventTypeOperationApproved:
				eventsSet.newApprovalEvents = append(eventsSet.newApprovalEvents,
					mustApprovalOpMsg(txInfo.tx.Hash.String(), cosmostypes.Event(event)).Message())
			case rarimotypes.EventTypeOperationRejected:
				eventsSet.newRejectionEvents = append(eventsSet.newRejectionEvents,
					mustRejectionOpMsg(txInfo.tx.Hash.String(), cosmostypes.Event(event)).Message())
			}
		}
	}

	return eventsSet
}

func mustToData(txs []txWithBlockInfo) []data.Transaction {
	res := make([]data.Transaction, len(txs))

	for i, txWithBlock := range txs {
		tx := *txWithBlock.tx

		txResultRaw, err := tx.TxResult.MarshalJSON()
		if err != nil {
			panic(errors.Wrap(err, "failed to marshal tx result", logan.F{
				"tx_hash": tx.Hash,
			}))
		}

		res[i] = data.Transaction{
			Hash: tx.Hash,
			BlockHeight: sql.NullInt64{
				Int64: tx.Height,
				Valid: true,
			},
			Index: sql.NullInt64{
				Int64: int64(tx.Index),
				Valid: true,
			},
			RawTx: []byte(tx.Tx.String()),
			TxResult: xo.NullJsonb{
				Jsonb: txResultRaw,
				Valid: true,
			},
			CreatedAt:   time.Now().UTC(),
			TxTimestamp: txWithBlock.blockTime,
		}
	}

	return res
}

func mustTransferOpMsg(txHash string, event cosmostypes.Event) msgs.TransferOpMsg {
	transferOpMsg := msgs.TransferOpMsg{
		TransactionHash: txHash,
	}

	for _, attr := range event.Attributes {
		if string(attr.Key) == rarimotypes.AttributeKeyOperationId {
			transferOpMsg.TransferID = string(attr.Value)
		}
	}

	if transferOpMsg.TransferID == "" {
		panic(fmt.Sprintf("no transfer id in event, tx=[%s]", txHash))
	}

	return transferOpMsg
}

func mustConfirmationOpMsg(txHash string, event cosmostypes.Event) msgs.ConfirmationOpMsg {
	confirmationOpMsg := msgs.ConfirmationOpMsg{
		TransactionHash: txHash,
	}

	for _, attr := range event.Attributes {
		if string(attr.Key) == rarimotypes.AttributeKeyConfirmationId {
			confirmationOpMsg.ConfirmationID = string(attr.Value)
		}
	}

	if confirmationOpMsg.ConfirmationID == "" {
		panic(fmt.Sprintf("no confirmation id in event, tx=[%s]", txHash))
	}

	return confirmationOpMsg
}

func mustVoteOpMsg(txHash string, event cosmostypes.Event) msgs.VoteOpMsg {
	voteOpMsg := msgs.VoteOpMsg{
		TransactionHash: txHash,
	}

	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case rarimotypes.AttributeKeyOperationId:
			voteOpMsg.OperationID = string(attr.Value)
		case rarimotypes.AttributeKeyOperationType:
			voteOpMsg.OperationType = string(attr.Value)
		case rarimotypes.AttributeKeyVotingChoice:
			voteOpMsg.VotingChoice = string(attr.Value)
		}
	}

	if voteOpMsg.OperationID == "" {
		panic(fmt.Sprintf("no operation id in event, tx=[%s]", txHash))
	}

	if voteOpMsg.OperationType == "" {
		panic(fmt.Sprintf("no operation type in event, tx=[%s]", txHash))
	}

	if voteOpMsg.VotingChoice == "" {
		panic(fmt.Sprintf("no voting choice in event, tx=[%s]", txHash))
	}

	return voteOpMsg
}

func mustApprovalOpMsg(txHash string, event cosmostypes.Event) msgs.ApprovalOpMsg {
	approvalOpMsg := msgs.ApprovalOpMsg{
		TransactionHash: txHash,
	}

	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case rarimotypes.AttributeKeyOperationId:
			approvalOpMsg.OperationID = string(attr.Value)
		case rarimotypes.AttributeKeyOperationType:
			approvalOpMsg.OperationType = string(attr.Value)
		}
	}

	if approvalOpMsg.OperationID == "" {
		panic(fmt.Sprintf("no operation id in event, tx=[%s]", txHash))
	}

	if approvalOpMsg.OperationType == "" {
		panic(fmt.Sprintf("no operation type in event, tx=[%s]", txHash))
	}

	return approvalOpMsg
}

func mustRejectionOpMsg(txHash string, event cosmostypes.Event) msgs.RejectionOpMsg {
	rejectionOpMsg := msgs.RejectionOpMsg{
		TransactionHash: txHash,
	}

	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case rarimotypes.AttributeKeyOperationId:
			rejectionOpMsg.OperationID = string(attr.Value)
		case rarimotypes.AttributeKeyOperationType:
			rejectionOpMsg.OperationType = string(attr.Value)
		}
	}

	if rejectionOpMsg.OperationID == "" {
		panic(fmt.Sprintf("no operation id in event, tx=[%s]", txHash))
	}

	if rejectionOpMsg.OperationType == "" {
		panic(fmt.Sprintf("no operation type in event, tx=[%s]", txHash))
	}

	return rejectionOpMsg
}

func rarimocoreBlockrangeCursor(start, end int64) string {
	return fmt.Sprintf("rarmocore-producer_block_range_cursor:%d-%d", start, end)
}

func isOutOfRangeError(err error) bool {
	if err == nil {
		return false
	}

	const outOfRangeCode = -32603 // TODO is it right or it's just internal error ?
	terr, ok := errors.Cause(err).(*types.RPCError)

	// yeah, i know, but there's no other way
	return ok && terr.Code == outOfRangeCode && strings.HasPrefix(terr.Data, "page should be within")
}
