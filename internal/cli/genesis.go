package cli

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/services"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func ParseAndSaveGenesis(ctx context.Context, cfg config.Config) {
	cfg.Log().Info("parsing genesis")

	err := services.NewTokenmanagerSaver(cfg).ParseAndSaveGenesis(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse and save genesis"))
	}

	cfg.Log().Info("genesis parsed and saved")
}
