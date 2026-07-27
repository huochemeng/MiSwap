package main

import (
	"MiSwap/config"
	"flag"
	"fmt"
)

const defaultConfigPath = "./config/config.toml"

func main() {
	fmt.Printf("go project init success")
	//	1.命令行参数与配置加载
	conf := flag.String("conf", defaultConfigPath, "config file path")
	//解析命令行参数。用户通过终端传入的配置路径赋值给conf
	flag.Parse()
	//*conf获取实际配置文件路径字符串，调用自定义函数将其反序列化成config结构体
	c, err := config.UnmarshalConfig(*conf)
	//错误处理
	if err != nil {
		panic(err)
	}
	//测验是否可以成功读取配置文件信息
	fmt.Printf("%v+\n", c)
	fmt.Printf("port%s", c.Api.Port)
}
