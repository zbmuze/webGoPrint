package utils

import (
	"fmt"
	"os"
)

// CreateDirIfNotExist 目录不存在则创建（含多级目录）
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm) // ModePerm：0777（所有权限）
	}
	return nil
}

// FormatFileSize ：格式化文件大小（B/KB/MB/GB）
func FormatFileSize(size int64) string {
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

// //GetUploadedFiles ：读取上传目录的所有文件并按时间排序（最新在前）
//func GetUploadedFiles() ([]models.FileInfo, error) {
//	var files []models.FileInfo
//
//	err := filepath.Walk(global.UploadDir, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//		if !info.IsDir() { // 只处理文件，跳过目录
//			files = append(files, models.FileInfo{
//				Name:       info.Name(),
//				Size:       FormatFileSize(info.Size()),
//				UploadTime: info.ModTime().Format("2006-01-02 15:04:05"),
//				Path:       path,
//			})
//		}
//		return nil
//	})
//
//	// 按上传时间降序排序
//	sort.Slice(files, func(i, j int) bool {
//		return files[i].UploadTime > files[j].UploadTime
//	})
//
//	return files, err
//}
