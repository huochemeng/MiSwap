package main

import (
	"MiSwap/api/router"
	"MiSwap/config"
	"MiSwap/service/svc"
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
	_ = c
	//	todo 服务上下文初始化。根据解析好的配置，创建服务上下文对象，包含数据库连接，缓存连接、日志记录等
	serverCtx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	// TODO 初始化路由，添加上下文对象注入到路由
	r := router.NewRouter(serverCtx)
	
	//启动
	r.Run()

}
