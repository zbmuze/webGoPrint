package utils

import (
	"errors"
	"log"
	"print-server/global"
	"print-server/models"
)

// AddQueueItem ：向数据库添加队列项（对应「文件上传」场景）
func AddQueueItem(item models.PrintQueue) error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()
	//fmt.Println(item) gorm会打印SQL语句
	res := global.DB.Create(&item)
	return res.Error
}

// GetErrorQueue 从数据库获取「打印失败」队列（对应「获取队列」场景）
func GetErrorQueue() ([]models.PrintQueue, error) {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	var fileInfos []models.PrintQueue
	// 直接查询并映射到 FileInfo 结构体
	result := global.DB.
		Model(&models.PrintQueue{}).
		Select("original_name as name, file_path as path, file_size as size, upload_time").
		Where("status = ?", "error").
		Order("upload_time DESC").
		Scan(&fileInfos)
	if result.Error != nil {
		return nil, result.Error
	}

	return fileInfos, nil
}

// GetWaitingQueue ：从数据库获取「待打印」队列（对应「获取队列」场景）
func GetWaitingQueue() ([]models.PrintQueue, error) {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	var fileInfos []models.PrintQueue
	// 直接查询并映射到 FileInfo 结构体
	result := global.DB.
		Model(&models.PrintQueue{}).                                                // 指定操作的模型
		Select("original_name", "file_path", "file_size", "upload_time", "status"). // 只查需要的字段
		Order("upload_time DESC").                                                  // 排序
		Find(&fileInfos)                                                            // 执行查询并将结果填充到 fileInfos
	if result.Error != nil {
		return nil, result.Error
	}
	return fileInfos, nil
}

// MarkItemPrinted ：将队列项标记为「已打印」（对应「单个文件打印」场景）
func MarkItemPrinted(originalName string) error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 先查询待打印的项是否存在（避免标记不存在的文件）
	var count int64
	// 检查符合条件的待打印记录是否存在
	err := global.DB.Model(&models.PrintQueue{}).
		Where("original_name = ? AND status = ?", originalName, "waiting").
		Count(&count).Error

	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("待打印文件不存在")
	}
	// 更新状态为已打印 'printed'
	result := global.DB.Model(&models.PrintQueue{}).
		Where("original_name = ? AND status = ?", originalName, "waiting").
		Update("status", "printed")
	//updateSQL := `UPDATE print_queue
	//			 SET status = 'printed'
	//			 WHERE original_name = ? AND status = 'waiting';`
	return result.Error
}

// UpdateItemStatus  更新状态为 status
//
//	filename： 文件名称
//	status：更新成什么状态
//	errorMsg :目前没用
//
// 返回值： 错误信息
func UpdateItemStatus(filename, status, errorMsg string) error {
	// 更新数据库中的状态
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 先查询待打印的项是否存在（避免标记不存在的文件）
	var count int64
	// 使用GORM查询符合条件的记录数
	err := global.DB.Model(models.PrintQueue{}).
		Where("original_name = ? AND status = ?", filename, "waiting").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("待打印文件不存在")
	}
	// 更新 original_name = filename 状态为 'printed'
	result := global.DB.Model(models.PrintQueue{}).
		Where("original_name = ? AND status = ?", filename, "waiting").
		Update("status", status)
	return result.Error
}

// MarkAllPrinted ：将所有「待打印」项标记为「已打印」（对应「打印全部」场景）
func MarkAllPrinted() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 批量更新状态
	result := global.DB.Model(&models.PrintQueue{}).
		Where("status = ?", "waiting").
		Update("status", "printed")
	return result.Error
}

// ClearWaitingQueue ：清空「待打印」队列（对应「清空队列」场景）
func ClearWaitingQueue() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	result := global.DB.Where("status = ?", "waiting").Delete(&models.PrintQueue{})
	return result.Error
}

// DeleteAllQueueItems ：删除所有队列项（对应「重置系统」场景）
func DeleteAllQueueItems() error {
	global.QueueMutex.Lock()
	defer global.QueueMutex.Unlock()

	// 删除所有记录（包括已打印和待打印）
	// 使用 Unscoped() 进行永久删除（硬删除）
	//result := global.DB.Unscoped().Where("1 = 1").Delete(&models.PrintQueue{})
	//if result.Error != nil {
	//	return result.Error
	//}
	// WHERE conditions required
	// 这个错误是因为 GORM 的安全机制导致的。当使用 Unscoped().Delete(&Model{}) 时，GORM 要求必须有 WHERE 条件，以防止意外删除所有数据。
	result := global.DB.Unscoped().Exec("DELETE FROM print_queue")
	if result.Error != nil {
		return result.Error
	}
	log.Printf("已删除所有队列记录，影响行数: %d", result.RowsAffected)
	return nil
}
