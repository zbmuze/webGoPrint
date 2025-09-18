package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"print-server/global"
	"print-server/models"
	"runtime"
	"sort"
	"strings"
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

// PrintDocument ：跨平台打印文件（Windows/macOS/Linux）
func PrintDocument(filePath string, printer string, pageSize string, orientation string) error {
	var cmd *exec.Cmd
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		cmd = exec.Command("print", filePath) // Windows打印命令
	case "darwin": // macOS
		cmd = exec.Command("lpr", filePath) // macOS打印命令
	default: // Linux
		//   -d <打印机名称>：指定要使用的打印机。
		//   -n <副本数>：指定打印份数。
		//   -o <选项>：指定打印选项，如双面打印、彩色打印等。
		//   -q <队列名称>：将打印任务添加到指定的打印队列。
		fmt.Printf("打印文件 %s,大小 %s，方向 %s 打印机 %s", filePath, pageSize, orientation, printer)
		cmd = exec.Command("lp", "-d", "Virtual_PDF_Printer", filePath) // Linux打印命令
	}
	return cmd.Run()
}

// GetUploadedFiles ：读取上传目录的所有文件并按时间排序（最新在前）
func GetUploadedFiles() ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.Walk(global.UploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() { // 只处理文件，跳过目录
			files = append(files, models.FileInfo{
				Name:       info.Name(),
				Size:       FormatFileSize(info.Size()),
				UploadTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Path:       path,
			})
		}
		return nil
	})

	// 按上传时间降序排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].UploadTime > files[j].UploadTime
	})

	return files, err
}

// AddQueueItem ：向数据库添加队列项（对应「文件上传」场景）
func AddQueueItem(item models.FileInfo) error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// SQL：插入队列项（status 默认为 'waiting'）
	sql := `INSERT INTO print_queue (original_name, file_path, file_size, upload_time)
			VALUES (?, ?, ?, ?);`

	// 执行插入（使用全局 DB 连接）
	_, err := global.DB.Exec(sql,
		item.Name,       // original_name
		item.Path,       // file_path
		item.Size,       // file_size
		item.UploadTime, // upload_time
	)
	return err
}

// GetWaitingQueue ：从数据库获取「待打印」队列（对应「获取队列」场景）
func GetWaitingQueue() ([]models.FileInfo, error) {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// SQL：查询 status = 'waiting' 的队列项，按 upload_time 降序（最新在前）
	sql := `SELECT original_name, file_path, file_size, upload_time 
			FROM print_queue 
			WHERE status = 'waiting' 
			ORDER BY upload_time DESC;`

	// 执行查询
	rows, err := global.DB.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // 确保查询结束后关闭结果集

	// 解析查询结果到切片
	var queue []models.FileInfo
	for rows.Next() {
		var item models.FileInfo
		err := rows.Scan(
			&item.Name,       // original_name -> FileInfo.Name
			&item.Path,       // file_path -> FileInfo.Path
			&item.Size,       // file_size -> FileInfo.Size
			&item.UploadTime, // upload_time -> FileInfo.UploadTime
		)
		if err != nil {
			return nil, err
		}
		queue = append(queue, item)
	}

	// 检查行迭代过程中的错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return queue, nil
}

// MarkItemPrinted ：将队列项标记为「已打印」（对应「单个文件打印」场景）
func MarkItemPrinted(originalName string) error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 先查询待打印的项是否存在（避免标记不存在的文件）
	var count int
	checkSQL := `SELECT COUNT(*) FROM print_queue 
				WHERE original_name = ? AND status = 'waiting';`
	err := global.DB.QueryRow(checkSQL, originalName).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("待打印文件不存在")
	}

	// SQL：更新状态为 'printed'
	updateSQL := `UPDATE print_queue 
				 SET status = 'printed' 
				 WHERE original_name = ? AND status = 'waiting';`
	_, err = global.DB.Exec(updateSQL, originalName)
	return err
}

// UpdateItemStatus  更新状态为打印失败
func UpdateItemStatus(filename, status, errorMsg string) error {
	// 更新数据库中的状态
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 先查询待打印的项是否存在（避免标记不存在的文件）
	var count int
	checkSQL := `SELECT COUNT(*) FROM print_queue 
				WHERE original_name = ? AND status = 'waiting';`
	err := global.DB.QueryRow(checkSQL, filename).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("待打印文件不存在")
	}

	// SQL：更新状态为 'printed'
	updateSQL := `UPDATE print_queue 
				 SET status = 'printed' 
				 WHERE original_name = ? AND status = 'waiting';`
	_, err = global.DB.Exec(updateSQL, filename)
	return err
}

// MarkAllPrinted ：将所有「待打印」项标记为「已打印」（对应「打印全部」场景）
func MarkAllPrinted() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// SQL：批量更新状态
	sql := `UPDATE print_queue 
			SET status = 'printed' 
			WHERE status = 'waiting';`
	_, err := global.DB.Exec(sql)
	return err
}

// ClearWaitingQueue ：清空「待打印」队列（对应「清空队列」场景）
func ClearWaitingQueue() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// SQL：删除所有待打印项（也可选择更新状态，根据需求）
	sql := `DELETE FROM print_queue WHERE status = 'waiting';`
	_, err := global.DB.Exec(sql)
	return err
}

// DeleteAllQueueItems ：删除所有队列项（对应「重置系统」场景）
func DeleteAllQueueItems() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// SQL：删除所有记录（包括已打印和待打印）
	sql := `DELETE FROM print_queue;`
	_, err := global.DB.Exec(sql)
	return err
}
