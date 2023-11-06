package types

import (
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
)

type UseMethod borsh.Enum

type Uses struct {
	UseMethod UseMethod
	Remaining uint64
	Total     uint64
}

type TokenStandard borsh.Enum

type MetaplexMetadata struct {
	Key                 borsh.Enum
	UpdateAuthority     solana.PublicKey
	Mint                solana.PublicKey
	Data                Data
	PrimarySaleHappened bool
	IsMutable           bool
	EditionNonce        *uint8
	TokenStandard       *TokenStandard
	Collection          *Collection
	Uses                *Uses
}

type Data struct {
	Name                 string
	Symbol               string
	URI                  string
	SellerFeeBasisPoints uint16
	Creators             *[]Creator
}

type Creator struct {
	Address  solana.PublicKey
	Verified bool
	Share    uint8
}

type Collection struct {
	Verified bool
	Address  solana.PublicKey
}
