package main

import (
	"flag"
	"fmt"
	"log"
	"print-server/global"
	routes "print-server/router"
	"print-server/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 解析命令行端口参数（默认8080）
	port := flag.Int("p", 8080, "指定服务器监听端口,默认8080")
	flag.Parse()
	global.ServerPort = strconv.Itoa(*port) // 同步到全局配置

	// 2. 获取本机IP（赋值到全局）
	ip, err := utils.GetLocalIP()
	if err != nil {
		log.Printf("获取本地IP失败: %v", err)
		global.ServerIP = "localhost"
	} else {
		global.ServerIP = ip
	}

	// 3. 创建上传目录（从全局配置读取目录名）
	if err := utils.CreateDirIfNotExist(global.UploadDir); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 4. 初始化Gin引擎 & 注册路由
	r := gin.Default()
	routes.InitRouter(r) // 路由逻辑抽离到routes包

	// 5. 启动服务器
	listenAddr := fmt.Sprintf("%s:%s", global.ServerIP, global.ServerPort)
	log.Printf("服务器已启动: http://%s", listenAddr)
	log.Fatal(r.Run(listenAddr))
}
