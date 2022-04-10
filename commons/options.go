package commons

type PaginationOptions struct {
	Limit  uint `json:"limit,omitempty"`
	Offset uint `json:"offset,omitempty"`
}

type SortOptions struct {
	Sort []string `json:"sort,omitempty"`
	//Descending bool     `json:"descending,omitempty"`
}
