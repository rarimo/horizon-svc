/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import (
	"encoding/json"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type ChainAttributes struct {
	// Address of bridge contract in specific chain
	BridgeContract string          `json:"bridge_contract"`
	ChainParams    json.RawMessage `json:"chain_params"`
	// Type of blockchain by supported wallets, APIs, etc.  Enum: - `evm` - `0` - `solana` - `1` - `near` - `2` - `other` - `3`
	ChainType tokenmanager.NetworkType `json:"chain_type"`
	// Link to network icon
	Icon *string `json:"icon,omitempty"`
	Name string  `json:"name"`
}
