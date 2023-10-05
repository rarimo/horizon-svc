package data

import "context"

//go:generate xo schema "postgres://postgres:postgres@localhost:5432/rarimo_horizon?sslmode=disable" -o ./ --single=schema.xo.go --src templates
//go:generate xo schema "postgres://postgres:postgres@localhost:5432/rarimo_horizon?sslmode=disable" -o pg --single=schema.xo.go --src=pg/templates --go-context=both
//go:generate goimports -w ./

type Storage interface {
	Transaction(func() error) error
	TransferQ() TransferQ
	ConfirmationQ() ConfirmationQ
	TransactionQ() TransactionQ
	VoteQ() VoteQ
	ApprovalQ() ApprovalQ
	RejectionQ() RejectionQ

	CollectionQ() CollectionQ
	CollectionChainMappingQ() CollectionChainMappingQ
	ItemQ() ItemQ
	ItemChainMappingQ() ItemChainMappingQ
}

type TransferQ interface {
	SelectCtx(ctx context.Context, selector TransferSelector) ([]Transfer, error)
	UpsertBatchCtx(ctx context.Context, transfers ...Transfer) error
	TransferByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*Transfer, error)
	SetStatusByIndexCtx(ctx context.Context, status int, indexes ...string) error
}

type ConfirmationQ interface {
	InsertBatchCtx(ctx context.Context, confirmations ...Confirmation) error
	ConfirmationsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]Confirmation, error)
}

type TransactionQ interface {
	InsertBatchCtx(ctx context.Context, transactions ...Transaction) error
}

type VoteQ interface {
	InsertBatchCtx(ctx context.Context, votes ...Vote) error
	VotesByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]Vote, error)
}

type ApprovalQ interface {
	InsertBatchCtx(ctx context.Context, approvals ...Approval) error
	ApprovalsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]Approval, error)
}

type RejectionQ interface {
	InsertBatchCtx(ctx context.Context, rejections ...Rejection) error
	RejectionsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]Rejection, error)
}

type CollectionQ interface {
	InsertCtx(ctx context.Context, c *Collection) error
	CollectionByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*Collection, error)
	DeleteCtx(ctx context.Context, c *Collection) error
}

type CollectionChainMappingQ interface {
	InsertBatchCtx(ctx context.Context, chainMappings ...CollectionChainMapping) error
	InsertCtx(ctx context.Context, ccm *CollectionChainMapping) error

	CollectionChainMappingByCollectionNetworkCtx(ctx context.Context, collection int64, network int, isForUpdate bool) (*CollectionChainMapping, error)

	UpsertCtx(ctx context.Context, ccm *CollectionChainMapping) error

	DeleteByCollectionCtx(ctx context.Context, collection int64) error
	DeleteCtx(ctx context.Context, ccm *CollectionChainMapping) error
}

type ItemQ interface {
	InsertCtx(ctx context.Context, i *Item) error
	ItemByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*Item, error)
	ItemByIDCtx(ctx context.Context, id int64, isForUpdate bool) (*Item, error)
	UpdateCtx(ctx context.Context, i *Item) error
	DeleteCtx(ctx context.Context, i *Item) error
}

type ItemChainMappingQ interface {
	InsertBatchCtx(ctx context.Context, chainMappings ...ItemChainMapping) error
	InsertCtx(ctx context.Context, icm *ItemChainMapping) error

	ItemChainMappingByItemNetworkCtx(ctx context.Context, item int64, network int, isForUpdate bool) (*ItemChainMapping, error)
	ItemChainMappingsByNetworkCtx(ctx context.Context, network int, isForUpdate bool) ([]ItemChainMapping, error)

	DeleteByItemCtx(ctx context.Context, item int64) error
	DeleteCtx(ctx context.Context, icm *ItemChainMapping) error
}

type SeedQ interface {
}

type GorpMigrationQ interface {
}
