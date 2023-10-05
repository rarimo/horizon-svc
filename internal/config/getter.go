package config

import (
	"fmt"
	"strings"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/kit/kv"

	"github.com/spf13/cast"
)

type pathGetter struct {
	Data kv.Getter
}

func newPathGetter(data kv.Getter) pathGetter {
	return pathGetter{
		Data: data,
	}
}

func (m pathGetter) GetStringMap(key string) (map[string]interface{}, error) {
	pathElemts := strings.Split(key, ".")
	path := ""
	data := m.Data
	for i, elem := range pathElemts {
		path = strings.TrimPrefix(path+"."+elem, ".")
		rawData, err := data.GetStringMap(elem)
		if err != nil {
			return nil, errors.From(err, logan.F{
				"path": path,
			})
		}

		if rawData == nil {
			return nil, nil
		}

		if i == len(pathElemts)-1 {
			return rawData, nil
		}

		data = mapGetter(rawData)
	}

	panic(fmt.Errorf("unexpected key: %s", key))
}

type mapGetter map[string]interface{}

func (m mapGetter) GetStringMap(key string) (map[string]interface{}, error) {
	raw, ok := m[key]
	if !ok {
		return nil, nil
	}

	result, err := cast.ToStringMapE(raw)
	if err != nil {
		return nil, err
	}

	return result, nil
}
