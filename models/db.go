package models

import (
	"time"

	"gorm.io/gorm"
)

// PrintQueue 定义打印队列模型
type PrintQueue struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	UploadTime   time.Time `gorm:"not null"`                             // 上传时间
	OriginalName string    `gorm:"type:TEXT;not null"`                   // 原名称
	FilePath     string    `gorm:"type:TEXT;not null;unique"`            // 文件路径
	FileSize     string    `gorm:"type:TEXT;not null"`                   // 文件大小
	Status       string    `gorm:"type:TEXT;not null;default:'waiting'"` // 文件状态
	Printer      string    `gorm:"type:TEXT;not null"`                   // 打印机选择
	PageSize     string    `gorm:"type:TEXT;not null;default:'A4"`       // 纸张大小 默认A4
	Orientation  string    `gorm:"type:TEXT;not null"`                   // 打印方向
}

func (PrintQueue) TableName() string {
	return "print_queue"
}
