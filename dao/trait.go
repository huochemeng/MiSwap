package dao

import (
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"context"
	"fmt"
	"github.com/pkg/errors"
)

func (d *Dao) QueryItemTraits(ctx context.Context, chain string, collectionAddr string, tokenID string) ([]model.ItemTrait, error) {
	var itemTraits []model.ItemTrait
	if err := d.DB.WithContext(ctx).Table(model.ItemTraitTableName(chain)).
		Select("collection_address, token_id, trait, trait_value").
		Where("collection_address = ? and token_id = ?", collectionAddr, tokenID).
		Scan(&itemTraits).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query items trait info")
	}

	return itemTraits, nil
}

// QueryCollectionTraits 查询NFT合集的 Trait信息统计
func (d *Dao) QueryCollectionTraits(ctx context.Context, chain string, collectionAddr string) ([]dto.TraitCount, error) {
	var traitCounts []dto.TraitCount
	if err := d.DB.WithContext(ctx).Table(model.ItemTraitTableName(chain)).
		Select("`trait`,`trait_value`,count(*) as count").Where("collection_address=?", collectionAddr).
		Group("`trait`,`trait_value`").
		Scan(&traitCounts).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query collection trait amount")
	}

	return traitCounts, nil
}

// QueryTraitsPrice 查询NFT Trait的价格信息
// 主要功能:
// 1. 查询指定NFT集合中特定token id的 Trait价格
// 2. 通过关联订单表和 Trait表,找出每个 Trait对应的最低挂单价格
// 3. 返回 Trait价格列表
func (d *Dao) QueryTraitsPrice(ctx context.Context, chain, collectionAddr string, tokenIds []string) ([]dto.TraitPrice, error) {
	var traitsPrice []dto.TraitPrice

	// 构建子查询,查询指定token的 Trait信息
	listSubQuery := d.DB.WithContext(ctx).
		Table(fmt.Sprintf("%s as gf_order", model.OrderTableName(chain))).
		// 查询字段: Trait名称、 Trait值、最低价格
		Select("gf_attribute.trait,gf_attribute.trait_value,min(gf_order.price) as price").
		// 条件1:匹配集合地址、订单类型为挂单、订单状态为活跃
		Where("gf_order.collection_address=? and gf_order.order_type=? and gf_order.order_status = ?",
			collectionAddr,
			model.ListingOrder,
			model.OrderStatusActive).
		// 条件2: Trait必须在指定token的 Trait列表中
		Where("(gf_attribute.trait,gf_attribute.trait_value) in (?)",
			d.DB.WithContext(ctx).
				Table(fmt.Sprintf("%s as gf_attr", model.ItemTraitTableName(chain))).
				Select("gf_attr.trait, gf_attr.trait_value").
				Where("gf_attr.collection_address=? and gf_attr.token_id in (?)",
					collectionAddr, tokenIds))

	// 关联 Trait表,按 Trait分组查询
	if err := listSubQuery.
		Joins(fmt.Sprintf("join %s as gf_attribute on gf_order.collection_address = gf_attribute.collection_address "+
			"and gf_order.token_id=gf_attribute.token_id", model.ItemTraitTableName(chain))).
		Group("gf_attribute.trait, gf_attribute.trait_value").
		Scan(&traitsPrice).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query trait price")
	}

	return traitsPrice, nil
}

// QueryItemsTraits 查询多个NFT Item的 Trait信息
func (d *Dao) QueryItemsTraits(ctx context.Context, chain string, collectionAddr string, tokenIds []string) ([]model.ItemTrait, error) {
	var itemsTraits []model.ItemTrait
	if err := d.DB.WithContext(ctx).Table(model.ItemTraitTableName(chain)).
		Select("collection_address, token_id, trait, trait_value").
		Where("collection_address = ? and token_id in (?)", collectionAddr, tokenIds).
		Scan(&itemsTraits).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query items trait info")
	}

	return itemsTraits, nil
}
