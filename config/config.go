package config

import (
	"MiSwap/base/stores/gdb"
	"github.com/spf13/viper"
	"strings"
)

type ProjectCfg struct {
	Name string `toml:"name" mapstructure:"name" json:"name"`
}

type Api struct {
	Port   string `toml:"port" mapstructure:"port" json:"port"`
	MaxNum int64  `toml:"max_num" json:"max_num"`
}

type Config struct {
	ProjectCfg *ProjectCfg `toml:"project_cfg" mapstructure:"project_cfg" json:"project_cfg"`
	Api        `toml:"api" json:"api"`
	DB         gdb.Config `toml:"db" json:"db"`
}

func UnmarshalConfig(configFilePath string) (*Config, error) {
	//	设置配置文件路径
	viper.SetConfigFile(configFilePath)
	//	设置配置文件格式
	viper.SetConfigType("toml")
	//	自动读取环境变量
	viper.AutomaticEnv()
	//	环境变量添加前缀
	viper.SetEnvPrefix("MI")
	//	设置替换器，把点号换成下划线
	replacer := strings.NewReplacer(".", "_")
	//	配置文件toml中使用点号进行层级划分，但是操作系统不支持点号，这里将点号换成下划线然后再环境变量查找
	viper.SetEnvKeyReplacer(replacer)

	//	读取配置文件内容
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	//初始化一个空的配置文件信息
	config, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	//	反序列化解析到config结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}
	//	返回配置信息结构体
	return config, nil

}

// DefaultConfig 默认函数，生成空配置结构信息
func DefaultConfig() (*Config, error) {
	return &Config{}, nil
}
