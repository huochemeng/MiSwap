package service

import (
	"MiSwap/base/pkg/cachekey"
	"MiSwap/base/pkg/crypto"
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func GetLoginMessage(ctx context.Context, svcCtx *svc.ServerCtx, address string) (*dto.UserLoginMsgResp, error) {
	nonce := uuid.NewString()
	loginMsg := fmt.Sprintf("Welcome to MiSwap!\n Nonce:%s", nonce)
	if err := svcCtx.KvStore.Setex(cachekey.GenerateLoginMsgCacheKey(address), nonce, 72*60*60); err != nil {
		return nil, errors.Wrap(err, "failed to generate login information")
	}
	return &dto.UserLoginMsgResp{Address: address, Message: loginMsg}, nil
}

func UserLogin(ctx context.Context, svcCtx *svc.ServerCtx, req dto.LoginReq) (*dto.UserLoginInfo, error) {
	resp := dto.UserLoginInfo{}
	//	从缓存中获取登录信息的uuid
	cachedUUID, err := svcCtx.KvStore.Get(cachekey.GenerateLoginMsgCacheKey(req.Address))
	if err != nil || cachedUUID == "" {
		return nil, errcode.ErrTokenExpire
	}
	//	从发送的消息中获取uuid
	// 1. 只按第一个匹配项分割，避免内容冲突
	idx := strings.Index(req.Message, "Nonce:")
	if idx == -1 {
		return nil, errcode.ErrInvalidMessageFormat // 区分格式错误和Token过期
	}

	// 2. 提取并清洗
	nonce := strings.TrimSpace(req.Message[idx+len("Nonce:"):])
	if nonce == "" {
		return nil, errcode.ErrInvalidMessageFormat
	}
	//比较
	if nonce != cachedUUID {
		return nil, errcode.ErrTokenExpire
	}
	//	查询用户基本信息
	var user model.User
	err = svcCtx.DB.WithContext(ctx).Select("id, address, is_allowed").Where("address = ?", req.Address).Find(&user).Error
	if err != nil {
		return nil, errcode.ErrCustom.WithMsg("failed to query user")
	}
	//	如果用户不存在，那么就创建新用户
	if user.Id == 0 {
		now := time.Now().UnixMilli()
		newUser := model.User{
			Address:    req.Address,
			IsAllowed:  false,
			IsSigned:   true,
			CreateTime: now,
			UpdateTime: now,
		}
		//	添加用户
		err := svcCtx.DB.WithContext(ctx).Create(&newUser).Error
		if err != nil {
			return nil, errors.Wrap(err, "failed create new user")
		}
		// 将新用户信息同步回外层变量
		user = newUser
	}
	//	生成用户token
	tokenKey := cachekey.GenerateLoginTokenCacheKey(req.Address)

	if err := svcCtx.KvStore.Setex(tokenKey, req.Address, 15*24*60*60); err != nil {
		return nil, err
	}
	//根据用户的tokenKey生成对应的token，用来返回给前端
	userToken, err := crypto.AesEncryptGCM([]byte(tokenKey), []byte(cachekey.LoginSalt))
	if err != nil {
		return nil, errcode.ErrCustom.WithMsg("failed get user token")
	}
	// 设置返回结果
	resp.Token = hex.EncodeToString(userToken)
	resp.IsAllowed = user.IsAllowed

	return &resp, err
}

// 封装到crypto包
//func AesEncryptGCM(data []byte, key []byte) ([]byte, error) {
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return nil, fmt.Errorf("new cipher failed: %w", err)
//	}
//
//	gcm, err := cipher.NewGCM(block)
//	if err != nil {
//		return nil, fmt.Errorf("new gcm failed: %w", err)
//	}
//
//	// Nonce 长度由 GCM 决定（通常 12 字节），不需要手动 padding
//	nonce := make([]byte, gcm.NonceSize())
//	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
//		return nil, fmt.Errorf("generate nonce failed: %w", err)
//	}
//
//	// Seal 会将 nonce + 密文 + tag 拼接返回
//	// 解密时用 Open 即可自动验证完整性
//	ciphertext := gcm.Seal(nonce, nonce, data, nil)
//	return ciphertext, nil
//}

func GetSigStatusMsg(ctx context.Context, svcCtx *svc.ServerCtx, addr string) (*dto.UserSignStatusResp, error) {
	isSigned, err := svcCtx.Dao.GetUserSigStatusMsg(ctx, addr)
	if err != nil {
		return nil, errcode.NewCustomErr("failed on get user sign status")
	}
	return &dto.UserSignStatusResp{IsSigned: isSigned}, nil
}
