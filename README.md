### go + gin 实现本地打印Web服务
##### 实现手机上传PDF打印

###### [开源地址](https://gitee.com/li0shang/webGoPrint)



打印队列需要存储「文件核心信息」和「队列状态」，表结构设计如下：

| 字段名        | 类型    | 说明                                         | 约束                       |
| ------------- | ------- | -------------------------------------------- | -------------------------- |
| id            | INTEGER | 唯一主键（自增）                             | PRIMARY KEY AUTOINCREMENT  |
| original_name | TEXT    | 文件原始名称（用户上传时的文件名）           | NOT NULL                   |
| file_path     | TEXT    | 文件本地保存路径（唯一，避免重复）           | NOT NULL UNIQUE            |
| file_size     | TEXT    | 格式化后的文件大小（如 "2.5MB"）             | NOT NULL                   |
| upload_time   | TEXT    | 上传时间（格式：2006-01-02 15:04:05）        | NOT NULL                   |
| status        | TEXT    | 队列状态（waiting：待打印；printed：已打印） | NOT NULL DEFAULT 'waiting' |