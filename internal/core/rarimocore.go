package core

import (
	"context"
	rarimocoretypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rarimocore struct {
	rc rarimocoretypes.QueryClient
}

var _ Rarimocore = &rarimocore{}

func (r *rarimocore) GetConfirmation(ctx context.Context, root string) (*rarimocoretypes.Confirmation, error) {
	resp, err := r.rc.Confirmation(ctx, &rarimocoretypes.QueryGetConfirmationRequest{
		Root: root,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get confirmation from core", logan.F{
			"root": root,
		})
	}

	return &resp.Confirmation, nil
}

func (r *rarimocore) GetOperation(ctx context.Context, index string) (*rarimocoretypes.Operation, error) {
	resp, err := r.rc.Operation(ctx, &rarimocoretypes.QueryGetOperationRequest{
		Index: index,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get operation from core", logan.F{
			"index": index,
		})
	}

	return &resp.Operation, nil
}
