package ipfs

import (
	"net/url"
	"reflect"

	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type IPFSer interface {
	IPFS() Gateway
}

type ipfser struct {
	getter kv.Getter
	once   comfig.Once
}

func NewIPFSer(getter kv.Getter) IPFSer {
	return &ipfser{
		getter: getter,
	}
}

type IpfsCredentials struct {
	GatewayURL    *url.URL `fig:"gateway_url,required"`
	ProjectSecret string   `fig:"project_secret,required"`
	ProjectID     string   `fig:"project_id,required"`
}

func (c *ipfser) IPFS() Gateway {
	return c.once.Do(func() interface{} {
		var config struct {
			Credentials []IpfsCredentials `fig:"credentials,required"`
		}

		err := figure.
			Out(&config).
			With(infuraCredentialsHook).
			From(kv.MustGetStringMap(c.getter, "ipfs_gateway")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure infura_ipfs_gateway"))
		}
		if len(config.Credentials) == 0 {
			panic(errors.New("at least one set of Infura IPFS Gateway credentials required"))
		}

		return NewIpfsGateway(config.Credentials)
	}).(Gateway)
}

var infuraCredentialsHook = figure.Hooks{
	"[]ipfs.IpfsCredentials": func(value interface{}) (reflect.Value, error) {
		slice, err := cast.ToSliceE(value)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected slice")
		}
		var result []IpfsCredentials
		for idx, el := range slice {
			raw, err := cast.ToStringMapE(el)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "expected object", logan.F{
					"index": idx,
				})
			}
			var credentials IpfsCredentials
			if err := figure.Out(&credentials).From(raw).Please(); err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to figure out credentials", logan.F{
					"index": idx,
				})
			}
			result = append(result, credentials)
		}
		return reflect.ValueOf(result), nil
	},
}
