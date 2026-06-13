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
	"syscall"  // ✅ 添加这个导入
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
    // 参数解析
    port := flag.Int("p", 8088, "指定服务器端口,默认8088")
    ip := flag.String("ip", "", "指定服务器IP地址,默认自动获取本机IP")
    global.Debug = flag.Bool("debug", false, "启用调试模式")
    flag.Parse()
    
    // 设置运行模式
    if *global.Debug {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }
    
    // 设置端口
    global.ServerPort = fmt.Sprintf(":%d", *port)
    
    // 设置IP地址
    if *ip != "" {
        log.Printf("使用指定的IP地址: %s", *ip)
        global.ServerIP = *ip
        // 如果指定了IP，需要同步设置IPv4和IPv6地址
        global.ServerIPv4 = *ip
        global.ServerIPv6 = *ip
    } else {
        log.Println("自动获取本机IP地址")
        ips, err := utils.GetAllIPs()
        if err != nil {
            log.Printf("获取IP失败: %v，使用localhost", err)
            global.ServerIP = "localhost"
            global.ServerIPv4 = "127.0.0.1"
            global.ServerIPv6 = "::1"
        } else {
            if len(ips) > 0 {
                global.ServerIP = ips[0]
            }
            if len(ips) > 1 {
                global.ServerIPv4 = ips[1]
            }
            if len(ips) > 2 {
                global.ServerIPv6 = ips[2]
            }
        }
    }
    
    // 创建上传目录
    if err := utils.CreateDirIfNotExist(global.UploadDir); err != nil {
        log.Fatalf("创建上传目录失败: %v", err)
    }
    
    // 初始化数据库
    if err := global.InitSQLite(); err != nil {
        log.Fatalf("初始化数据库失败：%v", err)
    }
    log.Println("数据库初始化成功")
    
    // 设置路由
    r := gin.Default()
    setupStaticFileService(r)
    setupTemplates(r)
    routes.InitRouter(r)
    
    // 启动服务器
    srv := &http.Server{
        Addr:    global.ServerPort,
        Handler: r,
        // 添加超时设置（可选）
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // 打印访问地址
    printAccessAddresses()
    
    // 启动服务（优雅关闭）
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("服务器启动失败: %v", err)
        }
    }()
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    
    log.Println("正在关闭服务器...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("服务器强制关闭:", err)
    }
    log.Println("服务器已退出")
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

// 打印访问地址
func printAccessAddresses() {
    log.Println("=" * 50)
    log.Println("✅ 服务器启动成功！")
    log.Println("=" * 50)
    
    if global.ServerIP != "" && global.ServerIP != "localhost" {
        if global.ServerIPv4 != "" && global.ServerIPv4 != "0.0.0.0" {
            log.Printf("📍 IPv4 访问地址: http://%s%s", global.ServerIPv4, global.ServerPort)
        }
        if global.ServerIPv6 != "" && global.ServerIPv6 != "::" {
            log.Printf("📍 IPv6 访问地址: http://[%s]%s", global.ServerIPv6, global.ServerPort)
        }
        log.Printf("📍 本机访问地址: http://localhost%s", global.ServerPort)
    } else {
        log.Printf("📍 访问地址: http://localhost%s", global.ServerPort)
    }
    log.Println("=" * 50)
    log.Println("按 Ctrl+C 停止服务器")
}
