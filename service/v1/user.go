package service

import (
	"MiSwap/base/pkg/cachekey"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func GetLoginMessage(ctx context.Context, svcCtx *svc.ServerCtx, address string) (*dto.UserLoginMsgResp, error) {
	nonce := uuid.NewString()
	loginMsg := fmt.Sprintf("Welcome to MiSwap!\n Nonce:%s", nonce)
	//	todo 缓存起来
	if err := svcCtx.KvStore.Setex(cachekey.GenerateLoginMsgCacheKey(address), nonce, 72*60*60); err != nil {
		return nil, errors.Wrap(err, "failed to generate login information")
	}
	return &dto.UserLoginMsgResp{Address: address, Message: loginMsg}, nil
}
