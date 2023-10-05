package rpc

type ResponsePaginatedResponse struct {
	TotalPages int `json:"totalPages,omitempty"`
	PageNumber int `json:"pageNumber,omitempty"`
	TotalItems int `json:"totalItems,omitempty"`
}
