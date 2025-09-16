package global

import "print-server/models"

// 上传目录（可直接修改）
var UploadDir = "uploads"

// 打印队列（存储待打印的文件信息）
var Queue []models.FileInfo

// 服务器IP&Port（启动时初始化）
var ServerIP string
var ServerPort string

// 支持的文件格式（可扩展）
var SupportedExt = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".txt": true,
}
