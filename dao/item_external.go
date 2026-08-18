package dao

import (
	"MiSwap/base/stores/gdb/model"
	"context"
	"github.com/pkg/errors"
)

// QueryCollectionItemsImage 查询集合内NFT Item的图片和视频信息
func (d *Dao) QueryCollectionItemsImage(ctx context.Context, chain string,
	collectionAddr string, tokenIds []string) ([]model.ItemExternal, error) {
	var itemsExternal []model.ItemExternal

	if err := d.DB.WithContext(ctx).
		Table(model.ItemExternalTableName(chain)).
		Select("collection_address, token_id, is_uploaded_oss, "+
			"image_uri, oss_uri, video_type, is_video_uploaded, "+
			"video_uri, video_oss_uri").
		Where("collection_address = ? and token_id in (?)",
			collectionAddr, tokenIds).
		Scan(&itemsExternal).Error; err != nil {
		return nil, errors.Wrap(err, "failed on query items external info")
	}

	return itemsExternal, nil
}
