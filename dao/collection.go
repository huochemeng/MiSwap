package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
)

var collectionDetailFields = []string{"id", "chain_id", "token_standard", "name", "address", "image_uri", "floor_price", "sale_price", "item_amount", "owner_amount"}

func (d *Dao) QueryCollectionInfo(ctx context.Context, chain string, addr string) (*model.Collection, error) {
	var collection model.Collection
	if err := d.DB.WithContext(ctx).Table(model.CollectionTableName(chain)).
		Select(collectionDetailFields).Where("address = ?", addr).
		First(&collection).Error; err != nil {
		return nil, errcode.NewCustomErr("dao: failed to get collection information")
	}
	return &collection, nil
}
