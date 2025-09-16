package handlers

import (
	"fmt"
	"net/http"
	"print-server/global"
	"print-server/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeHomePage：渲染PC端首页（含移动设备检测）
func ServeHomePage(c *gin.Context) {
	// 检测是否为移动设备（通过User-Agent）
	userAgent := c.Request.UserAgent()
	isMobile := strings.Contains(userAgent, "Mobile") ||
		strings.Contains(userAgent, "Android") ||
		strings.Contains(userAgent, "iPhone") ||
		strings.Contains(userAgent, "iPad")

	if isMobile {
		c.Redirect(http.StatusFound, "/mobile") // 移动设备跳转到移动端页面
		return
	}

	// PC端渲染首页（传递服务器地址）
	data := models.TemplateData{
		ServerAddress: fmt.Sprintf("%s:%s", global.ServerIP, global.ServerPort),
	}
	c.HTML(http.StatusOK, "index.html", data)
}

// ServeMobilePage：渲染移动端上传页面
func ServeMobilePage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}
