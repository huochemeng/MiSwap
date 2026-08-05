package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
)

func (d *Dao) GetUserSigStatusMsg(ctx context.Context, addr string) (bool, error) {
	var userInfo model.User
	err := d.DB.WithContext(ctx).Where("address = ?", addr).Find(&userInfo).Error
	if err != nil {
		return false, errcode.ErrCustom.WithMsg("failed on get user info")
	}
	return userInfo.IsSigned, nil
}
