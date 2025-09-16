package global

import (
	"database/sql"
	"sync"

	// _ "github.com/mattn/go-sqlite3" // SQLite 驱动（_ 表示只导入不直接使用）
	_ "modernc.org/sqlite"
)

var (

	// 上传目录（可直接修改）
	UploadDir    = "uploads"
	ServerIP     string
	ServerPort   string
	SupportedExt = map[string]bool{
		".pdf": true, ".doc": true, ".docx": true,
		".jpg": true, ".jpeg": true, ".png": true, ".txt": true,
	}
	QueueMutex sync.Mutex // 并发安全锁（同时保护数据库操作）
)

var DB *sql.DB                  // 全局数据库连接
const DBFile = "print_queue.db" // 数据库文件路径（项目根目录）
const CreateQueueTableSQL = `
CREATE TABLE IF NOT EXISTS print_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_name TEXT NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    file_size TEXT NOT NULL,
    upload_time TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'waiting'
);`

// InitSQLite：初始化 SQLite 数据库（创建文件+表）
func InitSQLite() error {
	// 打开数据库（不存在则自动创建文件）
	db, err := sql.Open("sqlite", DBFile)
	if err != nil {
		return err
	}

	// 验证连接（SQLite 是文件数据库，Open 不立即连接，需 Ping 确认）
	if err := db.Ping(); err != nil {
		return err
	}

	// 创建打印队列表
	_, err = db.Exec(CreateQueueTableSQL)
	if err != nil {
		return err
	}

	// 赋值全局连接（后续所有操作使用此连接）
	DB = db
	return nil
}
