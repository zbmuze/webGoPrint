package routes

import (
	"net/http"
	"print-server/global"
	"print-server/handlers"

	"github.com/gin-gonic/gin"
)

// InitRouter：初始化所有路由（接收Gin引擎，注册接口）
func InitRouter(r *gin.Engine) {
	// 1. 加载HTML模板（根目录下的templates文件夹）
	r.LoadHTMLFiles("templates/index.html", "templates/upload.html")

	// 2. 静态文件服务（静态资源、上传文件）
	r.Static("/static", "./static")                    // 静态资源（CSS/JS等）
	r.StaticFS("/uploads", http.Dir(global.UploadDir)) // 上传文件访问

	// 3. 页面路由（PC端首页、移动端页面）
	r.GET("/", handlers.ServeHomePage)         // PC端首页
	r.GET("/mobile", handlers.ServeMobilePage) // 移动端上传页

	// 4. 功能路由（二维码、文件上传、打印）
	r.GET("/qrcode", handlers.GenerateQRCode)   // 生成二维码
	r.GET("/queue", handlers.GetQueue)          // 获取打印队列
	r.POST("/upload", handlers.HandleUpload)    // 文件上传
	r.POST("/print", handlers.PrintFile)        // 打印单个文件
	r.POST("/print_all", handlers.PrintAll)     // 打印所有文件
	r.POST("/clear_queue", handlers.ClearQueue) // 清空队列
	r.POST("/reset", handlers.ResetSystem)      // 重置系统
}
