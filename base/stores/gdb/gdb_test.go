package gdb

import (
	"MiSwap/base/stores/gdb/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestConfig_CreateDataBase(t *testing.T) {
	c := &Config{
		User:     "root",
		Password: "123456",
		Host:     "localhost",
		Port:     3306,
		Database: "test_create_db",
	}
	assert.NoError(t, c.CreateDataBase())
}

func TestConfig_GetGormConfig(t *testing.T) {
	c := &Config{}
	assert.NotNil(t, c.GetGormConfig())
}

func setupTestDb(t *testing.T) *gorm.DB {
	t.Helper()
	c := &Config{
		User:     "root",
		Password: "123456",
		Host:     "localhost",
		Port:     3306,
		Database: "miswap",
	}
	err := c.CreateDataBase()
	if err != nil {
		t.Fatalf("failed CreateDataBase : %v", err)
	}
	db, err := NewDB(c)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	//	自动创建user表
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("failed to migrate user table: %v", err)
	}
	return db
}

func TestCreateUser(t *testing.T) {

	db := setupTestDb(t)
	user := model.User{
		Address:   "123456mini",
		IsAllowed: false,
	}

	result := db.Create(&user)
	assert.NoError(t, result.Error)
	assert.NotZero(t, user.Id)
}
