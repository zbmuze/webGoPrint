package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

// GetQueue 获取当前打印队列
func GetQueue(c *gin.Context) {
	queue, err := utils.GetWaitingQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取队列失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": queue})
}

// HandleUpload 处理文件上传（验证格式、大小，保存文件并加入队列）
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式（仅支持PDF/图片/TXT）"})
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

	isImage := global.ImageExts[ext]
	var pdfFilePath string
	if isImage {
		// 生成PDF文件名
		pdfFileName := uniqueFileName[:len(uniqueFileName)-len(ext)] + ".pdf"
		pdfFilePath = filepath.Join(global.UploadDir, pdfFileName)

		// 调用图片转PDF函数
		if err := utils.ConvertImageToPDF(filePath, pdfFilePath); err != nil {
			// 转换失败，清理已保存的图片文件并返回错误
			removeErr := os.Remove(filePath)
			if removeErr != nil {
				fmt.Printf("file remove err: %v\n", err)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("图片转换为PDF失败：%v", err)})
			return
		}
		// 转换成功，删除原始图片文件
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("删除原始图片文件失败: %v\n", err)
			// 这里不返回错误，因为PDF已经生成成功
		}
		// 更新文件路径为PDF路径
		filePath = pdfFilePath
		// 获取PDF文件大小
		if pdfInfo, err := os.Stat(pdfFilePath); err == nil {
			header.Size = pdfInfo.Size()
		}
	}

	fileInfo := models.PrintQueue{
		OriginalName: header.Filename, // 显示原始文件名
		FileSize:     utils.FormatFileSize(header.Size),
		UploadTime:   time.Now(),
		FilePath:     filePath,
		Printer:      global.Printer,
		PageSize:     global.PageSize,
		Orientation:  global.Orientation,
		Status:       "waiting", // 等待打印
	}
	// 8. 如果启用自动打印，立即执行打印
	if global.AutoPrint {
		// 新开线程
		go func() {
			// 延迟一下确保数据库事务完成
			time.Sleep(100 * time.Millisecond)
			if err := utils.PrintDocument(filePath); err != nil {
				fmt.Printf("自动打印失败: %v\n", err)
				// 更新状态为打印失败
				_ = utils.UpdateItemStatus(header.Filename, "failed", err.Error())
				return
			}
			// 打印成功，更新状态
			if err := utils.MarkItemPrinted(header.Filename); err != nil {
				fmt.Printf("更新打印状态失败: %v\n", err)
			}
		}()
	}

	if err := utils.AddQueueItem(fileInfo); err != nil {
		_ = os.Remove(filePath) // 数据库插入失败，删除已保存的文件
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加到队列失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "filename": header.Filename})
}

// PrintFile 打印单个文件（从队列或上传目录中查找）
func PrintFile(c *gin.Context) {
	var req struct {
		Filename    string `json:"filename"` // 前端传递的原始文件名
		Printer     string `json:"printer"`
		PageSize    string `json:"pageSize"`    // 大小 A4
		Orientation string `json:"orientation"` // 方向
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求参数"})
		return
	}
	// 从数据库获取文件路径（替代原从 global.Queue 查找）
	queue, err := utils.GetWaitingQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取队列失败：" + err.Error()})
		return
	}

	var filePath string
	for _, file := range queue {
		if file.OriginalName == req.Filename {
			filePath = file.FilePath
			break
		}
	}

	if filePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件未找到：" + req.Filename})
		return
	}

	// 检查文件是否存在（原有逻辑不变）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 新增：文件不存在时，从数据库删除该队列项
		err := utils.MarkItemPrinted(req.Filename)
		if err != nil {
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "文件已被删除"})
		return
	}

	// 执行打印命令
	if err := utils.PrintDocument(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败：" + err.Error()})
		return
	}

	if err := utils.MarkItemPrinted(req.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新队列状态失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "打印任务已发送"})
}

// PrintAll 打印队列中所有文件（打印后清空队列）
func PrintAll(c *gin.Context) {
	queue, err := utils.GetWaitingQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取队列失败：" + err.Error()})
		return
	}
	if len(queue) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "打印队列为空"})
		return
	}

	// 打印所有文件（若单个失败则返回错误）
	for _, file := range queue {
		if err := utils.PrintDocument(file.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败：" + err.Error()})
			return
		}
	}
	// 标记所有为已打印
	if err := utils.MarkAllPrinted(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新队列状态失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有打印任务已发送"})
}

// ClearQueue 清空打印队列（不删除文件）
func ClearQueue(c *gin.Context) {
	if err := utils.ClearWaitingQueue(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空队列失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "打印队列已清空"})
}

// ResetSystem 重置系统（清空队列+删除所有上传文件）
func ResetSystem(c *gin.Context) {
	// 1. 清空数据库所有队列项（新增）
	if err := utils.DeleteAllQueueItems(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空队列失败：" + err.Error()})
		return
	}
	// 2. 删除上传目录文件（原有逻辑不变）
	if err := filepath.Walk(global.UploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置系统失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "系统已重置"})
}
