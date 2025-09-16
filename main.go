package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// 文件信息结构
type FileInfo struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	UploadTime string `json:"upload_time"`
	Path       string `json:"-"`
}

// 模板数据
type TemplateData struct {
	ServerAddress string
}

// 全局变量
var (
	uploadDir    = "uploads"
	queue        []FileInfo
	serverIP     string
	serverPort   = "8080"
	supportedExt = map[string]bool{
		".pdf": true, ".doc": true, ".docx": true,
		".jpg": true, ".jpeg": true, ".png": true, ".txt": true,
	}
)

func main() {

	// 定义命令行参数，默认端口为8080
	port := flag.Int("p", 8080, "指定服务器监听端口，默认8080")
	flag.Parse() // 解析命令行参数

	// 获取本机IP地址
	ip, err := getLocalIP()
	if err != nil {
		log.Printf("获取本地IP失败: %v", err)
		serverIP = "localhost"
	} else {
		serverIP = ip
	}

	// 创建上传目录
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 初始化Gin
	r := gin.Default()

	// 加载HTML模板
	// 加载HTML模板
	r.LoadHTMLFiles("templates/index.html", "templates/upload.html")

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFS("/uploads", http.Dir(uploadDir))

	// API路由
	r.GET("/", serveHomePage)
	r.GET("/mobile", serveMobilePage)
	r.GET("/qrcode", generateQRCode)
	r.GET("/queue", getQueue)
	r.POST("/upload", handleUpload)
	r.POST("/print", printFile)
	r.POST("/print_all", printAll)
	r.POST("/clear_queue", clearQueue)
	r.POST("/reset", resetSystem)

	// 启动服务器
	// log.Printf("服务器启动: http://%s:%s", serverIP, serverPort)
	// 启动服务器，使用命令行指定的端口
	listenAddr := fmt.Sprintf("%s:%d", serverIP, *port)
	log.Printf("服务器将在 http://%s 启动...\n", listenAddr)
	log.Fatal(r.Run(listenAddr))
}

// 添加移动端页面服务函数
func serveMobilePage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

// 更新 serveHomePage 函数，添加移动设备检测
func serveHomePage(c *gin.Context) {
	// 检测是否为移动设备
	userAgent := c.Request.UserAgent()
	isMobile := strings.Contains(userAgent, "Mobile") ||
		strings.Contains(userAgent, "Android") ||
		strings.Contains(userAgent, "iPhone") ||
		strings.Contains(userAgent, "iPad")

	if isMobile {
		// 如果是移动设备，重定向到移动端页面
		c.Redirect(http.StatusFound, "/mobile")
		return
	}

	data := TemplateData{
		ServerAddress: fmt.Sprintf("%s:%s", serverIP, serverPort),
	}
	c.HTML(http.StatusOK, "index.html", data)
}

// 生成二维码
func generateQRCode(c *gin.Context) {
	url := fmt.Sprintf("http://%s:%s/mobile", serverIP, serverPort)
	qr, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "生成二维码失败")
		return
	}

	c.Data(http.StatusOK, "image/png", qr)
}

// 获取打印队列
func getQueue(c *gin.Context) {
	// 读取上传目录中的文件
	files, err := getUploadedFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// 处理文件上传
func handleUpload(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	defer file.Close()

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !supportedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
		return
	}

	// 检查文件大小（最大10MB）
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过10MB限制"})
		return
	}

	// 生成唯一文件名
	hash := md5.Sum([]byte(header.Filename + time.Now().String()))
	filename := hex.EncodeToString(hash[:]) + ext
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 添加到队列
	fileInfo := FileInfo{
		Name:       header.Filename,
		Size:       formatFileSize(header.Size),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		Path:       filePath,
	}

	queue = append(queue, fileInfo)

	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "filename": header.Filename})
}

// 在 printFile 函数中，改进文件查找逻辑
func printFile(c *gin.Context) {
	var request struct {
		Filename string `json:"filename"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}

	// 首先检查文件是否在队列中
	var filePath string
	var fileIndex = -1

	for i, file := range queue {
		if file.Name == request.Filename {
			filePath = file.Path
			fileIndex = i
			break
		}
	}

	if filePath == "" {
		// 如果不在队列中，检查上传目录中是否有该文件
		files, _ := getUploadedFiles()
		for _, file := range files {
			if file.Name == request.Filename {
				filePath = file.Path
				// 添加到队列
				queue = append(queue, file)
				fileIndex = len(queue) - 1
				break
			}
		}

		if filePath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件未找到: " + request.Filename})
			return
		}
	}

	// 检查文件是否实际存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 从队列中移除不存在的文件
		if fileIndex >= 0 {
			queue = append(queue[:fileIndex], queue[fileIndex+1:]...)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "文件已被删除"})
		return
	}

	// 打印文件
	if err := printDocument(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败: " + err.Error()})
		return
	}

	// 从队列中移除文件
	if fileIndex >= 0 {
		queue = append(queue[:fileIndex], queue[fileIndex+1:]...)
	}

	c.JSON(http.StatusOK, gin.H{"message": "打印任务已发送"})
}

// 打印全部文件
func printAll(c *gin.Context) {
	if len(queue) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "打印队列为空"})
		return
	}

	// 打印所有文件
	for _, file := range queue {
		if err := printDocument(file.Path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打印失败: " + err.Error()})
			return
		}
	}

	// 清空队列
	queue = []FileInfo{}

	c.JSON(http.StatusOK, gin.H{"message": "所有打印任务已发送"})
}

// 清空队列
func clearQueue(c *gin.Context) {
	queue = []FileInfo{}
	c.JSON(http.StatusOK, gin.H{"message": "打印队列已清空"})
}

// 重置系统
func resetSystem(c *gin.Context) {
	// 清空队列
	queue = []FileInfo{}

	// 删除上传目录中的所有文件
	err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return os.Remove(path)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置系统失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "系统已重置"})
}

// 辅助函数：获取本地IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("未找到有效IP地址")
}

// 辅助函数：获取已上传的文件列表
func getUploadedFiles() ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, FileInfo{
				Name:       info.Name(),
				Size:       formatFileSize(info.Size()),
				UploadTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Path:       path,
			})
		}
		return nil
	})

	// 按上传时间排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		return files[i].UploadTime > files[j].UploadTime
	})

	return files, err
}

// 辅助函数：格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 辅助函数：打印文档
func printDocument(filePath string) error {
	// 根据系统选择打印命令
	var cmd *exec.Cmd
	switch os := strings.ToLower(runtime.GOOS); os {
	case "windows":
		cmd = exec.Command("print", filePath)
	case "darwin": // macOS
		cmd = exec.Command("lpr", filePath)
	default: // Linux
		cmd = exec.Command("lp", filePath)
	}

	return cmd.Run()
}
