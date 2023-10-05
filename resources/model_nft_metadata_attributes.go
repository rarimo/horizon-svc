/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type NftMetadataAttributes struct {
	AnimationUrl *string                `json:"animation_url,omitempty"`
	Attributes   []NftMetadataAttribute `json:"attributes"`
	Description  *string                `json:"description,omitempty"`
	ExternalUrl  *string                `json:"external_url,omitempty"`
	// Link to image
	ImageUrl string `json:"image_url"`
	// original url to metadata stored in the contract
	MetadataUrl string `json:"metadata_url"`
	Name        string `json:"name"`
}
