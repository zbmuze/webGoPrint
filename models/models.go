package models

// FileInfo 文件信息结构（用于队列、文件列表展示）
type FileInfo struct {
	Name        string `json:"name"`        // 文件名（原始名）
	Size        string `json:"size"`        // 文件大小（格式化后，如"2.5MB"）
	UploadTime  string `json:"upload_time"` // 上传时间（格式化后）
	Path        string `json:"path"`        // 文件本地路径（不对外返回）
	Printer     string `json:"printer"`
	PageSize    string `json:"page_size"`
	Orientation string `json:"orientation"`
	Status      string `json:"status"` // 等待打印
}
