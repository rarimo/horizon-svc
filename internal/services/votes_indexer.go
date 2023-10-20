package services

import (
	"context"
	"time"

	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"

	"github.com/rarimo/horizon-svc/pkg/msgs"

	"gitlab.com/distributed_lab/logan/v3"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
)

func RunVotesIndexer(ctx context.Context, cfg config.Config) {
	vindexer := &votesIndexer{
		log:     cfg.Log().WithField("who", cfg.VotesIndexer().RunnerName),
		storage: cfg.NewStorage().Clone(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.VotesIndexer().VotesConsumer,
		vindexer,
	).Run(ctx)
}

type votesIndexer struct {
	log     *logan.Entry
	storage data.Storage
}

func (p *votesIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	votes := make([]data.Vote, len(msgs))

	for i, msg := range msgs {
		vmsg := msg.MustVoteOpMessage()

		vote := data.Vote{
			TransferIndex:     []byte(vmsg.OperationID),
			RarimoTransaction: data.MustDBHash(vmsg.TransactionHash),
			CreatedAt:         time.Now().UTC(),
		}

		switch vmsg.VotingChoice {
		case rarimocore.VoteType_YES.String():
			vote.Choice = int(rarimocore.VoteType_YES)
		case rarimocore.VoteType_NO.String():
			vote.Choice = int(rarimocore.VoteType_NO)
		default:
			p.log.WithFields(logan.F{
				"voting_choice": vmsg.VotingChoice,
				"operation_id":  vmsg.OperationID,
			}).Warn("unknown voting choice received, skipping vote")
		}

		votes[i] = vote
	}

	return p.storage.VoteQ().InsertBatchCtx(ctx, votes...)
}
