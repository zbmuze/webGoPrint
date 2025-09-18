package handlers

import (
	"net/http"
	"print-server/global"

	"github.com/gin-gonic/gin"
)

func Setting(c *gin.Context) {
	var req struct {
		Printer     string `json:"printer"`     // 打印机
		PageSize    string `json:"pageSize"`    // 大小 A4
		Orientation string `json:"orientation"` // 方向
		IsAutoPrint bool   `json:"isauto"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求参数"})
		return
	}
	global.Printer = req.Printer
	global.PageSize = req.PageSize
	global.Orientation = req.Orientation
	global.AutoPrint = req.IsAutoPrint
	c.JSON(http.StatusOK, gin.H{"ok": req})
}
