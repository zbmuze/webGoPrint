package handlers

import (
	"fmt"
	"net/http"
	"print-server/global"

	"github.com/gin-gonic/gin"
)

// TemplateData 模板渲染数据（用于首页HTML传递参数）
type TemplateData struct {
	ServerAddress string // 服务器地址（IP:Port）
}

// ServeHomePage 渲染服务端首页
func ServeHomePage(c *gin.Context) {
	// PC端渲染首页（传递服务器地址）
	data := TemplateData{
		ServerAddress: fmt.Sprintf("%s:%s", global.ServerIP, global.ServerPort),
	}
	c.HTML(http.StatusOK, "index.html", data)
}

// ServeMobilePage 渲染移动端上传页面
func ServeMobilePage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}
