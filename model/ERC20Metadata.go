package model

// ERC20Metadata is the definition of the Token meta info
type ERC20Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"owner"`
	TotalSupply uint64 `json:"totalsupply"`
}

// newERC20Metadata is ...
func newERC20Metadata(name, symbol, owner string, totalSupply uint64) *ERC20Metadata {
	return &ERC20Metadata{name, symbol, owner, totalSupply}
}
