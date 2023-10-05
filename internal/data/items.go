package data

type ItemsSelector struct {
	Network    *string `json:"network,omitempty"`
	Collection *int64  `json:"collection,omitempty"`
}
