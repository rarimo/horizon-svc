package metadata_fetcher

import (
	"encoding/base64"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/url"
	"strings"
)

func parseDataContentType(u url.URL) ([]byte, error) {
	mimeTypeData := strings.Split(u.Opaque, ";")
	if len(mimeTypeData) != 2 {
		return nil, errors.New("unexpected number of data url parts")
	}

	mimeType := mimeTypeData[0]
	if mimeType != "application/json" {
		return nil, errors.From(errors.New("unexpected mime type"), logan.F{
			"mime_type": mimeType,
		})
	}

	dataPart := mimeTypeData[1]
	encodingPayload := strings.Split(dataPart, ",")
	if len(encodingPayload) != 2 {
		return nil, errors.New("unexpected format of data url's payload")
	}

	encoding := encodingPayload[0]
	if encoding != "base64" {
		return nil, errors.New("expected base64 encoding")
	}

	payload, err := base64.StdEncoding.DecodeString(encodingPayload[1])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode payload")
	}

	return payload, nil
}
