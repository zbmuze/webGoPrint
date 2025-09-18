package handlers

import (
	"print-server/utils"

	"github.com/gin-gonic/gin"
)

// GetPrinters 获取打印机列表
func GetPrinters(c *gin.Context) {
	printers, err := utils.GetPrinters()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取打印机列表失败"})
		return
	}
	c.JSON(200, printers)
}
