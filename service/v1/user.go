package service

import (
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"fmt"
	"github.com/google/uuid"
)

func GetLoginMessage(ctx context.Context, svcCtx *svc.ServerCtx, address string) (*dto.UserLoginMsgResp, error) {
	nonce := uuid.NewString()
	loginMsg := fmt.Sprintf("Welcome to MiSwap!\n Nonce:%s", nonce)
	//	todo 缓存起来
	return &dto.UserLoginMsgResp{Address: address, Message: loginMsg}, nil
}
