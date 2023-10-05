package txbuild

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rarimo/evm-bridge-contracts/gobind/contracts/bridge"
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

type MultiBuilder struct {
	chains data.ChainsQ

	evmBuild    *ethtx.Builder
	solanaBuild *soltx.Builder
	nearBuild   *neartx.Builder
}

func NewMultiBuilder(cfg config.Config) *MultiBuilder {
	abi, err := bridge.BridgeMetaData.GetAbi()
	if err != nil {
		panic(errors.Wrap(err, "failed to get bridge abi"))
	}

	return &MultiBuilder{
		chains:      cfg.ChainsQ(),
		evmBuild:    ethtx.NewBuilder(cfg.EVM().RPCClient, abi, cfg.EVM().BridgeContract),
		solanaBuild: soltx.NewBuilder(cfg.Solana().RPCClient, cfg.Solana().ProgramID, cfg.Solana().BridgeAdminPublicKey),
		nearBuild:   neartx.NewBuilder(cfg.Near().Client, cfg.Near().BridgeAddress),
	}
}

func (t *MultiBuilder) BuildTx(ctx context.Context, req *resources.BuildTx, txData interface{}) (*resources.UnsubmittedTx, error) {
	net := t.chains.Get(req.Attributes.Network)
	if net == nil {
		return nil, ErrUnsupportedNetworkType
	}

	switch net.Type {
	case tokenmanager.NetworkType_EVM:
		return t.evmBuild.BuildTx(ctx, req, txData)
	case tokenmanager.NetworkType_Solana:
		return t.solanaBuild.BuildTx(ctx, req, txData)
	case tokenmanager.NetworkType_Near:
		return t.nearBuild.BuildTx(ctx, req, txData)
	default:
		panic(fmt.Errorf("unsupported network type %s", net.Type))
	}
}
