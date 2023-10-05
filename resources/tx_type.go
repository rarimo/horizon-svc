package resources

import (
	"encoding/json"
	"strconv"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type TxType int

const (
	TxTypeDepositNative TxType = iota
	TxTypeDepositErc20
	TxTypeDepositErc721
	TxTypeDepositErc1155
	TxTypeDepositFT
	TxTypeDepositNFT
)

var txTypeIntStr = map[TxType]string{
	TxTypeDepositNative:  "deposit_native",
	TxTypeDepositErc20:   "deposit_erc20",
	TxTypeDepositErc721:  "deposit_erc721",
	TxTypeDepositErc1155: "deposit_erc1155",
	TxTypeDepositFT:      "deposit_ft",
	TxTypeDepositNFT:     "deposit_nft",
}

var txTypeStrInt = map[string]TxType{
	"deposit_native":  TxTypeDepositNative,
	"deposit_erc20":   TxTypeDepositErc20,
	"deposit_erc721":  TxTypeDepositErc721,
	"deposit_erc1155": TxTypeDepositErc1155,
	"deposit_ft":      TxTypeDepositFT,
	"deposit_nft":     TxTypeDepositNFT,
}

func (t TxType) String() string {
	return txTypeIntStr[t]
}

func (t TxType) MarshalJSON() ([]byte, error) {
	return json.Marshal(Flag{
		Name:  txTypeIntStr[t],
		Value: int32(t),
	})
}

func (t *TxType) UnmarshalJSON(b []byte) error {
	var res Flag
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*t = TxType(res.Value)
	return nil
}

func (t *TxType) UnmarshalText(b []byte) error {
	typ, err := strconv.ParseInt(string(b), 0, 0)
	if err != nil {
		return err
	}

	if _, ok := txTypeIntStr[TxType(typ)]; !ok {
		return errors.From(errors.New("unsupported value"), logan.F{
			"supported": []TxType{
				TxTypeDepositNative,
				TxTypeDepositErc20,
				TxTypeDepositErc721,
				TxTypeDepositErc1155,
			},
		})
	}

	*t = TxType(typ)
	return nil
}

func SupportedTxTypesText() []string {
	return []string{
		TxTypeDepositNative.String(),
		TxTypeDepositErc20.String(),
		TxTypeDepositErc721.String(),
		TxTypeDepositErc1155.String(),
		TxTypeDepositFT.String(),
		TxTypeDepositNFT.String(),
	}
}

func SupportedTxTypesSolana() []interface{} {
	return []interface{}{
		TxTypeDepositNative,
		TxTypeDepositFT,
		TxTypeDepositNFT,
	}
}

func SupportedTxTypesNear() []interface{} {
	return []interface{}{
		TxTypeDepositNative,
		TxTypeDepositFT,
		TxTypeDepositNFT,
	}
}

func SupportedTxTypesEth() []interface{} {
	return []interface{}{
		TxTypeDepositNative,
		TxTypeDepositErc20,
		TxTypeDepositErc721,
		TxTypeDepositErc1155,
	}
}

func SupportedTxTypes() []interface{} {
	return []interface{}{
		TxTypeDepositNative,
		TxTypeDepositErc20,
		TxTypeDepositErc721,
		TxTypeDepositErc1155,
		TxTypeDepositFT,
		TxTypeDepositNFT,
	}
}

func (t TxType) Int() int {
	return int(t)
}
