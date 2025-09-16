package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
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

	// 2. 初始化 SQLite 数据库（新增：必须在路由初始化前执行）
	if err := global.InitSQLite(); err != nil {
		log.Fatalf("初始化数据库失败：%v", err)
	}
	log.Println("SQLite 数据库初始化成功")
	// 3. 获取本机IP（赋值到全局）
	ip, err := utils.GetLocalIP()
	if err != nil {
		log.Printf("获取本地IP失败: %v", err)
		global.ServerIP = "localhost"
	} else {
		global.ServerIP = ip
	}

	// 4. 创建上传目录（从全局配置读取目录名）
	if err := utils.CreateDirIfNotExist(global.UploadDir); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 5. 初始化Gin引擎 & 注册路由
	r := gin.Default()
	// 设置静态文件服务
	setupStaticFileService(r)
	// 设置模板
	setupTemplates(r)
	routes.InitRouter(r) // 路由逻辑抽离到routes包

	// 6. 启动服务器
	listenAddr := fmt.Sprintf("%s:%s", global.ServerIP, global.ServerPort)
	log.Printf("服务器已启动: http://%s", listenAddr)
	log.Fatal(r.Run(listenAddr))
}

//go:embed templates/* static/*
var embeddedFiles embed.FS

// 设置静态文件服务
func setupStaticFileService(router *gin.Engine) {
	// 使用 fs.Sub 获取 static 子目录的文件系统
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem for static files: %v", err)
	}

	// 提供静态文件服务
	router.StaticFS("/static", http.FS(staticFS))
}

// 设置模板
func setupTemplates(router *gin.Engine) {
	// 使用 fs.Sub 获取 templates 子目录的文件系统
	templateFS, err := fs.Sub(embeddedFiles, "templates")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem for templates: %v", err)
	}

	// 加载模板
	tmpl := template.Must(template.New("").ParseFS(templateFS, "*.html"))
	router.SetHTMLTemplate(tmpl)
}
