package data

import (
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type Token struct {
	ID     string  `fig:"id,required"`
	Name   string  `fig:"name,required"`
	Symbol string  `fig:"symbol,required"`
	Icon   *string `fig:"icon"`
	// Relation
	Chains []TokenChain `fig:"chains,required"`
}

type TokenChain struct {
	TokenID         string            `fig:"token_id,required"`
	TokenType       tokenmanager.Type `fig:"token_type,required"`
	ChainID         string            `fig:"chain_id,required"`
	ContractAddress string            `fig:"contract_address,required"`
	Wrapped         bool              `fig:"wrapped"`
}
