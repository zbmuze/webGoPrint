package routes

import (
	"net/http"
	"print-server/global"
	"print-server/handlers"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化所有路由
func InitRouter(r *gin.Engine) {

	// 上传文件从本地目录提供访问
	r.StaticFS("/uploads", http.Dir(global.UploadDir))

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

	r.GET("/get_printers", handlers.GetPrinters) // 获取打印机
	// r.POST("/merge", handlers.Merge) // 合并PDF
}
