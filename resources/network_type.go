package resources

import (
	"encoding/json"
	"strconv"

	"github.com/rarimo/rarimo-core/x/tokenmanager/types"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type NetworkType types.NetworkType

const (
	NetworkTypeEVM          = NetworkType(types.NetworkType_EVM)
	NetworkTypeSolana       = NetworkType(types.NetworkType_Solana)
	NetworkTypeNearProtocol = NetworkType(types.NetworkType_Near)
	NetworkTypeOther        = NetworkType(types.NetworkType_Other)
)

var networkTypeIntStr = map[NetworkType]string{
	NetworkTypeEVM:          "evm",
	NetworkTypeSolana:       "solana",
	NetworkTypeNearProtocol: "nearprotocol",
	NetworkTypeOther:        "other",
}

var networkTypeStrInt = map[string]NetworkType{
	"evm":          NetworkTypeEVM,
	"solana":       NetworkTypeSolana,
	"nearprotocol": NetworkTypeNearProtocol,
	"other":        NetworkTypeOther,
}

func (t NetworkType) String() string {
	return networkTypeIntStr[t]
}

func (t NetworkType) MarshalJSON() ([]byte, error) {
	return json.Marshal(Flag{
		Name:  networkTypeIntStr[t],
		Value: int32(t),
	})
}

func (t *NetworkType) UnmarshalJSON(b []byte) error {
	var res Flag
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*t = NetworkType(res.Value)
	return nil
}

func SupportedNetworkTypesText() []string {
	return []string{
		NetworkTypeEVM.String(),
		NetworkTypeSolana.String(),
		NetworkTypeNearProtocol.String(),
	}
}

func SupportedNetworkTypes() []interface{} {
	return []interface{}{
		NetworkTypeEVM,
		NetworkTypeSolana,
		NetworkTypeNearProtocol,
	}
}

func (t *NetworkType) UnmarshalText(b []byte) error {
	typ, err := strconv.ParseInt(string(b), 0, 0)
	if err != nil {
		return err
	}

	if _, ok := networkTypeIntStr[NetworkType(typ)]; !ok {
		return errors.From(errors.New("unsupported value"), logan.F{
			"supported": SupportedNetworkTypes(),
		})
	}

	*t = NetworkType(typ)
	return nil
}

func (t NetworkType) Int() int {
	return int(t)
}
