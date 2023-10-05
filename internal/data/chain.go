package data

import (
	"encoding/json"

	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type ChainsQ interface {
	Get(chain string) *Chain
	List() []Chain
	Page(pageNum, limit int) []Chain
}

type Chain struct {
	ID          int                      `fig:"id,required"`
	Name        string                   `fig:"name,required"`
	Rpc         string                   `fig:"rpc,required"`
	Type        tokenmanager.NetworkType `fig:"type,required"`
	Icon        *string                  `fig:"icon"`
	ChainParams json.RawMessage          `fig:"chain_params"`
}
