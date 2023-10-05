package txbuild

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/near-go/nearclient"

	"github.com/pkg/errors"
	gobind "github.com/rarimo/evm-bridge-contracts/gobind/contracts/bridge"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/ethtx"
	"github.com/rarimo/horizon-svc/pkg/neartx"
	"github.com/rarimo/horizon-svc/pkg/soltx"
	"github.com/rarimo/horizon-svc/resources"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
)

var (
	ErrUnsupportedNetworkType = errors.New("unsupported network type")
)

type Builder interface {
	BuildTx(ctx context.Context, req *resources.BuildTx, rawTxData interface{}) (*resources.UnsubmittedTx, error)
}

type MultiBuilder struct {
	chains   data.ChainsQ
	builders map[string]Builder
}

func NewMultiBuilder(cfg config.Config) *MultiBuilder {
	abi, err := gobind.BridgeMetaData.GetAbi()
	if err != nil {
		panic(errors.Wrap(err, "failed to get bridge abi"))
	}

	networks, err := cfg.Core().TokenManager().GetParams(context.Background())
	if err != nil {
		panic(errors.Wrap(err, "failed to get networks"))
	}

	builders := make(map[string]Builder)
	contracts := make(map[string]bridgeConfig)

	for _, net := range networks.Networks {
		bridgeParams := net.GetBridgeParams()
		if bridgeParams == nil {
			continue
		}

		contracts[net.Name] = bridgeConfig{
			contract: bridgeParams.Contract,
			//adminPublicKey: bridgeParams.AdminPublicKey,
			// TODO: ADD ADMIN PUBLIC KEY THERE
		}
	}

	for _, chain := range cfg.ChainsQ().List() {
		chainConf := contracts[chain.Name]
		if chainConf.contract == "" {
			panic(fmt.Errorf("contract address not found for chain %s", chain.Name))
		}

		switch chain.Type {
		case tokenmanager.NetworkType_EVM:
			cli, err := ethclient.Dial(chain.Rpc)
			if err != nil {
				panic(errors.Wrap(err, "failed to dial eth client"))
			}

			builders[chain.Name] = ethtx.NewBuilder(cli, abi, common.HexToAddress(chainConf.contract))
		case tokenmanager.NetworkType_Solana:
			if chainConf.adminPublicKey == "" {
				panic(fmt.Errorf("adminPublicKey not found for chain %s", chain.Name))
			}

			builders[chain.Name] = soltx.NewBuilder(
				rpc.New(chain.Rpc),
				solana.PublicKeyFromBytes(hexutil.MustDecode(chainConf.contract)),
				solana.PublicKeyFromBytes(hexutil.MustDecode(chainConf.adminPublicKey)),
			)
		case tokenmanager.NetworkType_Near:
			cli, err := nearclient.New(chain.Rpc)
			if err != nil {
				panic(errors.Wrap(err, "failed to dial near client"))
			}
			builders[chain.Name] = neartx.NewBuilder(cli, string(hexutil.MustDecode(chainConf.contract)))
		default:
			panic(fmt.Errorf("unsupported network type %s", chain.Type))
		}
	}

	return &MultiBuilder{
		cfg.ChainsQ(),
		builders,
	}
}

func (t *MultiBuilder) BuildTx(ctx context.Context, req *resources.BuildTx, txData interface{}) (*resources.UnsubmittedTx, error) {
	net := t.chains.Get(req.Attributes.Network)
	if net == nil {
		return nil, ErrUnsupportedNetworkType
	}

	return t.builders[net.Name].BuildTx(ctx, req, txData)
}

type bridgeConfig struct {
	adminPublicKey string
	contract       string
}
