package models

import (
	"time"

	"gorm.io/gorm"
)

// PrintQueue 定义打印队列模型
type PrintQueue struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	OriginalName string    `gorm:"type:TEXT;not null"`
	FilePath     string    `gorm:"type:TEXT;not null;unique"`
	FileSize     string    `gorm:"type:TEXT;not null"`
	UploadTime   time.Time `gorm:"not null"`
	Status       string    `gorm:"type:TEXT;not null;default:'waiting'"`
	Printer      string    `gorm:"type:TEXT;not null"`
	PageSize     string    `gorm:"type:TEXT;not null"`
	Orientation  string    `gorm:"type:TEXT;not null"`
}

func (PrintQueue) TableName() string {
	return "print_queue"
}
