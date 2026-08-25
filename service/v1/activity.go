package service

import (
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"github.com/pkg/errors"
)

func GetMultiChainActivities(ctx context.Context, svcCtx *svc.ServerCtx, chainID []int, chainName []string, addresses []string, tokenID string, userAddresses []string, eventTypes []string, page int, size int) (interface{}, interface{}) {
	activities, total, err := svcCtx.Dao.QueryMultiChainActivities(ctx, chainName, addresses, tokenID, userAddresses, eventTypes, page, size)
	if err != nil {
		return nil, errors.Wrap(err, "failed on query multi-chain activity")
	}

	if total == 0 || len(activities) == 0 {
		return &dto.ActivityResp{
			Result: nil,
			Count:  0,
		}, nil
	}

	//external info query
	results, err := svcCtx.Dao.QueryMultiChainActivityExternalInfo(ctx, chainID, chainName, activities)
	if err != nil {
		return nil, errors.Wrap(err, "failed on query activity external info")
	}

	return &dto.ActivityResp{
		Result: results,
		Count:  total,
	}, nil
}
