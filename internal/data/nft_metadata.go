package data

type NftMetadata struct {
	Name         string         `json:"name"`
	ImageURL     string         `json:"image_url"`
	MetadataUrl  *string        `json:"metadata_url,omitempty"`
	Description  *string        `json:"description,omitempty"`
	AnimationUrl *string        `json:"animation_url,omitempty"`
	ExternalUrl  *string        `json:"external_url,omitempty"`
	Attributes   []NftAttribute `json:"attributes,omitempty"`
}

type NftAttribute struct {
	Trait string `json:"trait_type"`
	Value string `json:"value"`
}

func (m *NftMetadata) Merge(metadata *NftMetadata) {
	m.MetadataUrl = GetOrDefaultStrPtr(m.MetadataUrl, metadata.MetadataUrl)
	m.Name = GetOrDefaultStr(m.Name, metadata.Name)
	m.ImageURL = GetOrDefaultStr(m.ImageURL, metadata.ImageURL)
	m.Description = GetOrDefaultStrPtr(m.Description, metadata.Description)
	m.AnimationUrl = GetOrDefaultStrPtr(m.AnimationUrl, metadata.AnimationUrl)
	m.ExternalUrl = GetOrDefaultStrPtr(m.ExternalUrl, metadata.ExternalUrl)

	if m.Attributes == nil || len(m.Attributes) == 0 {
		m.Attributes = metadata.Attributes
	}
}
