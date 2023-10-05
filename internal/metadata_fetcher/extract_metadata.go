package metadata_fetcher

import (
	"encoding/json"
	"fmt"
	"github.com/rarimo/horizon-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
)

var metadataNamesSynonyms = []string{"name", "title"}

func getName(metadata map[string]interface{}) string {
	return getStr(metadataNamesSynonyms, metadata)
}

var metadataImageURLSynonyms = []string{"image", "image_url_cdn", "image_url", "image_uri", "media"}

func getImageURL(metadata map[string]interface{}) string {
	return getStr(metadataImageURLSynonyms, metadata)
}

var metadataDescriptionURLSynonyms = []string{"description", "bio"}

func getDescription(metadata map[string]interface{}) string {
	return getStr(metadataDescriptionURLSynonyms, metadata)
}

var metadataAnimationURLSynonyms = []string{"animation", "animation_url"}

func getAnimationURL(metadata map[string]interface{}) string {
	return getStr(metadataAnimationURLSynonyms, metadata)
}

var metadataExternalURLSynonyms = []string{"external", "external_url"}

func getExternalURL(metadata map[string]interface{}) string {
	return getStr(metadataExternalURLSynonyms, metadata)
}

func getAttributes(metadata map[string]interface{}) []data.NftAttribute {
	for _, key := range []string{"attributes", "properties"} {
		result, ok := metadata[key]
		if !ok {
			continue
		}

		attributes, ok := result.([]interface{})
		if !ok {
			continue
		}

		attrs := make([]data.NftAttribute, 0, len(attributes))
		for _, attr := range attributes {
			traitType := getAttribute("trait_type", attr.(map[string]interface{}))
			value := getAttribute("value", attr.(map[string]interface{}))

			if traitType == "" || value == "" {
				continue
			}
			attrs = append(attrs, data.NftAttribute{
				Trait: traitType,
				Value: value,
			})
		}

		return attrs
	}

	return []data.NftAttribute{}
}

func getAttribute(key string, attr map[string]interface{}) string {
	result, ok := attr[key]
	if !ok {
		return ""
	}

	strResult, ok := result.(string)
	if ok && strResult != "" {
		return strResult
	}

	intResult, err := strconv.ParseInt(fmt.Sprintf("%v", result), 10, 64)
	if err == nil {
		return fmt.Sprintf("%d", intResult)
	}

	return ""
}

func getStr(synonyms []string, metadata map[string]interface{}) string {
	for _, key := range synonyms {
		result, ok := metadata[key]
		if !ok {
			continue
		}

		strResult, ok := result.(string)
		if ok {
			return strResult
		}
	}

	return ""
}

func extractMetadata(payload []byte, uri string) (*data.NftMetadata, error) {
	payloadAsObject := map[string]interface{}{}
	err := json.Unmarshal(payload, &payloadAsObject)
	if err != nil {
		return nil, errors.Wrap(err, "nft metadata is not an object", logan.F{
			"uri": uri,
		})
	}

	var metadata data.NftMetadata

	rawName := getName(payloadAsObject)
	if rawName != "" {
		metadata.Name = rawName
	}

	image := getImageURL(payloadAsObject)
	if image != "" {
		metadata.ImageURL = image
	}

	description := getDescription(payloadAsObject)
	if description != "" {
		metadata.Description = &description
	}

	animationURL := getAnimationURL(payloadAsObject)
	if animationURL != "" {
		metadata.AnimationUrl = &animationURL
	}

	externalURL := getExternalURL(payloadAsObject)
	if externalURL != "" {
		metadata.ExternalUrl = &externalURL
	}

	attributes := getAttributes(payloadAsObject)
	if len(attributes) != 0 {
		metadata.Attributes = attributes
	}

	metadata.MetadataUrl = &uri

	return &metadata, nil
}
