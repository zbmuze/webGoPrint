package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"print-server/global"
	"print-server/models"
	"print-server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetQueue：获取当前打印队列
func GetQueue(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"files": global.Queue})
}

// HandleUpload：处理文件上传（验证格式、大小，保存文件并加入队列）
func HandleUpload(c *gin.Context) {
	// 1. 获取上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	defer file.Close()

	// 2. 验证文件格式（对比全局支持的扩展名）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !global.SupportedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式（仅支持PDF/Word/图片/TXT）"})
		return
	}

	// 3. 验证文件大小（最大10MB）
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过10MB限制"})
		return
	}

	// 4. 生成唯一文件名（避免重复）
	hash := md5.Sum([]byte(header.Filename + time.Now().String()))
	uniqueFileName := hex.EncodeToString(hash[:]) + ext
	filePath := filepath.Join(global.UploadDir, uniqueFileName)

	// 5. 保存文件到本地
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 6. 将文件加入打印队列
	fileInfo := models.FileInfo{
		Name:       header.Filename, // 显示原始文件名
		Size:       utils.FormatFileSize(header.Size),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		Path:       filePath,
	}
	global.Queue = append(global.Queue, fileInfo)

	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "filename": header.Filename})
}

// PrintFile：打印单个文件（从队列或上传目录中查找）
func PrintFile(c *gin.Context) {
	var req struct {
		Filename string `json:"filename"` // 前端传递的原始文件名
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求参数"})
		return
	}

	// 1. 从队列中查找文件
	var filePath string
	var fileIndex = -1
	for i, file := range global.Queue {
		if file.Name == req.Filename {
			filePath = file.Path
			fileIndex = i
			break
		}
	}

	// 2. 队列中无文件时，从上传目录查找（并加入队列）
	if filePath == "" {
		uploadedFiles, _ := utils.GetUploadedFiles()
		for _, file := range uploadedFiles {
			if file.Name == req.Filename {
				filePath = file.Path
				global.Queue = append(global.Queue, file)
				fileIndex = len(global.Queue) - 1
				break
			}
		}
		if filePath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件未找到：" + req.Filename})
			return
		}
	}

	// 3. 检查文件是否实际存在（避免已删除的情况）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if fileIndex >= 0 {
			global.Queue = append(global.Queue[:fileIndex], global.Queue[fileIndex+1:]...) // 从队列移除
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "文件已被删除"})
		return
	}

	// 4. 执行打印命令
	if err := utils.PrintDocument(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败：" + err.Error()})
		return
	}

	// 5. 打印成功后从队列移除
	if fileIndex >= 0 {
		global.Queue = append(global.Queue[:fileIndex], global.Queue[fileIndex+1:]...)
	}

	c.JSON(http.StatusOK, gin.H{"message": "打印任务已发送"})
}

// PrintAll：打印队列中所有文件（打印后清空队列）
func PrintAll(c *gin.Context) {
	if len(global.Queue) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "打印队列为空"})
		return
	}

	// 打印所有文件（若单个失败则返回错误）
	for _, file := range global.Queue {
		if err := utils.PrintDocument(file.Path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败：" + err.Error()})
			return
		}
	}

	// 清空队列
	global.Queue = []models.FileInfo{}
	c.JSON(http.StatusOK, gin.H{"message": "所有打印任务已发送"})
}

// ClearQueue：清空打印队列（不删除文件）
func ClearQueue(c *gin.Context) {
	global.Queue = []models.FileInfo{}
	c.JSON(http.StatusOK, gin.H{"message": "打印队列已清空"})
}

// ResetSystem：重置系统（清空队列+删除所有上传文件）
func ResetSystem(c *gin.Context) {
	// 1. 清空打印队列
	global.Queue = []models.FileInfo{}

	// 2. 删除上传目录所有文件（保留目录本身）
	err := filepath.Walk(global.UploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() { // 只删除文件，不删除目录
			return os.Remove(path)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统重置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "系统已重置"})
}
