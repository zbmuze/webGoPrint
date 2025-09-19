package global

import (
	"fmt"
	"log"
	"print-server/models"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// UploadDir 上传目录（可直接修改）
	UploadDir    = "uploads"
	Debug        *bool
	ServerIP     string
	ServerIPv4   string
	ServerIPv6   string
	ServerPort   string
	SupportedExt = map[string]bool{
		".pdf": true, ".doc": true, ".docx": true,
		".jpg": true, ".jpeg": true, ".png": true, ".txt": true,
	}
	ImageExts = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}
	QueueMutex sync.Mutex // 并发安全锁（同时保护数据库操作）
	PrintMutex sync.Mutex // 打印机控制
)

var (
	// DB 全局数据库连接
	DB *gorm.DB
	// Printer 打印机默认
	Printer     string
	PageSize    string
	Orientation string
	AutoPrint   bool
)

const DBFile = "print.db" // 数据库文件路径（项目根目录）

// InitSQLite 初始化 SQLite 数据库（创建文件+表）
func InitSQLite() error {
	// 设置日志模式
	var logmode logger.LogLevel
	if *Debug {
		log.Printf("数据库：调试模式 \n")
		logmode = logger.Info
	} else {
		logmode = logger.Error
	}
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(DBFile), &gorm.Config{
		Logger: logger.Default.LogMode(logmode), // 设置日志级别，方便查看 SQL
	})
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	// 获取底层 SQL DB 连接进行 Ping 测试
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("无法获取 sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping数据库失败: %w", err)
	}
	// 配置连接池
	sqlDB.SetMaxOpenConns(1)            // SQLite 推荐最大连接数为 1
	sqlDB.SetMaxIdleConns(1)            // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	fmt.Println("数据库连接成功建立！")
	// 自动迁移模型（创建表）
	err = db.AutoMigrate(&models.PrintQueue{})
	if err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}
	// 赋值全局连接
	DB = db
	fmt.Println("使用GORM成功初始化SQLite数据库")
	return nil
}
