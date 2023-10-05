package data

import (
	"encoding/json"
	"time"

	"github.com/tendermint/tendermint/libs/bytes"

	"gitlab.com/distributed_lab/kit/pgdb"
)

type TransferSelector struct {
	Origin           *string    `json:"origin,omitempty"`
	RarimoTx         *string    `json:"bridge_tx,omitempty"`
	ChainTx          *string    `json:"original_tx,omitempty"`
	SourceChain      *string    `json:"src_chain,omitempty"`
	DestinationChain *string    `json:"dst_chain,omitempty"`
	Receiver         *string    `json:"receiver,omitempty"`
	Status           *int       `json:"signed,omitempty"`
	Creator          *string    `json:"creator,omitempty"`
	Before           *time.Time `json:"bridged_before,omitempty"`
	After            *time.Time `json:"bridged_after,omitempty"`
	TokenIndex       *string    `json:"token_index,omitempty"`

	PageCursor uint64     `json:"page_number,omitempty"`
	PageSize   uint64     `json:"page_size,omitempty"`
	Sort       pgdb.Sorts `json:"sort"`
}

func (s TransferSelector) MustCacheKey() string {
	key, err := json.Marshal(s)
	if err != nil {
		panic("failed to marshal transfer selector to json")
	}
	return "transfers_select:" + string(key)
}

func (s Transfer) RarimoTxHash() string {
	return bytes.HexBytes(s.RarimoTx).String()
}
