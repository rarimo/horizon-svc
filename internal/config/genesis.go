package config

import (
	"encoding/json"
	"os"
	"reflect"

	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GenesisCfg struct {
	Disabled     bool
	GenesisState *tokenmanager.GenesisState
}

func (c *config) Genesis() GenesisCfg {
	return c.genesis.Do(func() interface{} {
		var cfg struct {
			File     string `fig:"file"`
			Disabled bool   `fig:"disabled"`
		}

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "genesis")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out genesis config"))
		}

		if cfg.File == "" {
			return GenesisCfg{
				Disabled: cfg.Disabled,
			}
		}

		f, err := os.Open(cfg.File)
		if err != nil {
			panic(errors.Wrap(err, "failed to open genesis file", logan.F{
				"file": cfg.File,
			}))
		}

		defer f.Close()

		var genesis struct {
			AppState struct {
				TokenManager tokenmanager.GenesisState `json:"tokenmanager"`
			} `json:"app_state"`
		}

		if err := json.NewDecoder(f).Decode(&genesis); err != nil {
			panic(errors.Wrap(err, "failed to decode genesis file", logan.F{
				"file": cfg.File,
			}))
		}

		if reflect.DeepEqual(genesis.AppState.TokenManager, tokenmanager.GenesisState{}) {
			panic(errors.New("read genesis state is empty"))
		}

		return GenesisCfg{
			Disabled:     cfg.Disabled,
			GenesisState: &genesis.AppState.TokenManager,
		}
	}).(GenesisCfg)
}
