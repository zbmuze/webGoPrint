package handlers

import (
	"fmt"
	"net/http"
	"print-server/global"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// GenerateQRCode 生成移动端页面的二维码（地址：http://IP:Port/mobile）
func GenerateQRCode(c *gin.Context) {
	// 二维码指向的移动端页面地址
	mobileURL := fmt.Sprintf("http://%s:%s/mobile", global.ServerIP, global.ServerPort)
	// 生成二维码（Medium级别容错，256x256尺寸）
	qrData, err := qrcode.Encode(mobileURL, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "生成二维码失败")
		return
	}
	// 返回二维码图片（MIME类型为image/png）
	c.Data(http.StatusOK, "image/png", qrData)
}
