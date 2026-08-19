package dto

import "MiSwap/base/stores/gdb/model"

type TraitCount struct {
	model.ItemTrait
	Count int64 `json:"count"`
}

type ItemTraitsResp struct {
	Result interface{} `json:"result"`
}

type TraitInfo struct {
	Trait        string  `json:"trait"`
	TraitValue   string  `json:"trait_value"`
	TraitAmount  int64   `json:"trait_amount"`
	TraitPercent float64 `json:"trait_percent"`
}

type TraitValue struct {
	TraitValue  string `json:"trait_value"`
	TraitAmount int64  `json:"trait_amount"`
}

type CollectionTraitInfo struct {
	Trait  string       `json:"trait"`
	Values []TraitValue `json:"values"`
}
