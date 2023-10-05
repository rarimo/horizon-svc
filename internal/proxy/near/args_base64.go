package near

import (
	"encoding/base64"
	"encoding/json"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func argsToBase64(args map[string]interface{}) (*string, error) {
	params, err := json.Marshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal args to base64")
	}

	result := base64.StdEncoding.EncodeToString(params)
	return &result, nil
}
