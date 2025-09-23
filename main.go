package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"print-server/global"
	routes "print-server/router"
	"print-server/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 默认端口 8080 可以 -p 设置端口
	port := flag.Int("p", 8080, "指定服务器监听端口,默认8080")
	// 默认 release模式 -debug 可查看
	global.Debug = flag.Bool("debug", false, "启用调试模式")
	flag.Parse()
	if *global.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// 同步到全局配置
	global.ServerPort = ":" + strconv.Itoa(*port)
	// 获取本机IP（赋值到全局） GetAllIPInfo 获取公网IP
	ips, err := utils.GetAllIPs()
	//log.Printf("获取本地IPs: %v", ips)
	if err != nil {
		log.Printf("获取本地IP失败: %v", err)
		global.ServerIP = "localhost"
	} else {
		global.ServerIP = ips[0]
		global.ServerIPv4 = ips[1]
		global.ServerIPv6 = ips[2]
	}
	// 创建上传目录（从全局配置读取目录名）
	if err := utils.CreateDirIfNotExist(global.UploadDir); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
	// 初始化 SQLite 数据库（必须在路由初始化前执行）
	if err := global.InitSQLite(); err != nil {
		log.Fatalf("初始化数据库失败：%v", err)
	}
	log.Println("SQLite 数据库初始化成功")
	// 初始化Gin引擎
	r := gin.Default()
	// 设置静态文件服务
	setupStaticFileService(r)
	// 设置模板
	setupTemplates(r)
	// 注册路由
	routes.InitRouter(r)
	// 启动服务器

	// 创建自定义的HTTP服务器
	srv := &http.Server{
		Addr:    global.ServerPort,
		Handler: r,
	}
	listener, err := net.Listen("tcp", global.ServerPort)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	listenAddr := fmt.Sprintf("%s%s", global.ServerIP, global.ServerPort)
	listenAddrv4 := fmt.Sprintf("%s%s", global.ServerIPv4, global.ServerPort)
	listenAddrv6 := fmt.Sprintf("[%s]%s", global.ServerIPv6, global.ServerPort)
	log.Printf("服务器已启动 局域网: http://%s", listenAddr)
	log.Printf("服务器已启动 IPv4: http://%s", listenAddrv4)
	log.Printf("服务器已启动 IPv6: http://%s", listenAddrv6)
	// 在一个新的goroutine中启动服务器
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭：等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit // 阻塞，直到收到中断信号

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")

	// log.Fatal(r.Run(listenAddr))
}

//go:embed templates/* static/*
var embeddedFiles embed.FS

// 设置静态文件服务
func setupStaticFileService(router *gin.Engine) {
	// 使用 fs.Sub 获取 static 子目录的文件系统
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("为静态文件创建子文件系统失败: %v", err)
	}
	// 提供静态文件服务
	router.StaticFS("/static", http.FS(staticFS))
	// 2. 单独配置favicon.ico（浏览器默认请求路径）
	router.GET("/favicon.ico", func(c *gin.Context) {
		// 从嵌入的文件中读取static/favicon.ico
		data, err := embeddedFiles.ReadFile("static/favicon.ico")
		if err != nil {
			// 若文件不存在，返回404
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		// 设置正确的MIME类型（必须指定，否则浏览器无法识别）
		c.Data(http.StatusOK, "image/x-icon", data)
	})
}

// 设置模板
func setupTemplates(router *gin.Engine) {
	// 使用 fs.Sub 获取 templates 子目录的文件系统
	templateFS, err := fs.Sub(embeddedFiles, "templates")
	if err != nil {
		log.Fatalf("未能为模板创建子文件系统: %v", err)
	}
	// 加载模板
	tmpl := template.Must(template.New("").ParseFS(templateFS, "*.html"))
	router.SetHTMLTemplate(tmpl)
}
