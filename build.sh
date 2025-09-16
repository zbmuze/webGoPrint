#!/bin/bash

# 脚本功能：一键编译 Go 程序到 Linux amd64 和 armv7 架构
# 使用说明：1. 给脚本加执行权限：chmod +x build.sh
#          2. 执行脚本：./build.sh


# -------------------------- 配置参数 --------------------------
# 输出目录（编译后的文件会放在这里）
OUTPUT_DIR="build"
# 程序名称前缀（不同架构会自动加后缀）
APP_NAME="print_server"
# Go 编译额外参数（可选：如减小体积的 -ldflags "-s -w"）
GO_BUILD_FLAGS="-ldflags \"-s -w\""  # -s 去除符号表，-w 去除调试信息，可减小 30%+ 体积


# -------------------------- 函数定义 --------------------------
# 编译函数：参数1=目标架构(GOARCH)，参数2=目标ARM版本(GOARM，仅arm架构需要)，参数3=输出文件名后缀
build() {
    local arch=$1
    local goarm=$2
    local suffix=$3
    local output_path="${OUTPUT_DIR}/${APP_NAME}_${suffix}"

    # 打印编译状态
    echo -e "\033[32m[开始编译] Linux ${arch}（${suffix}）架构，输出路径：${output_path}\033[0m"

    # 设置编译环境变量
    export GOOS="linux"
    export GOARCH="${arch}"
    export CGO_ENABLED="0"
    if [ -n "${goarm}" ]; then  # 仅 arm 架构需要设置 GOARM
        export GOARM="${goarm}"
    fi

    # 执行编译命令（带错误捕获）
    go build ${GO_BUILD_FLAGS} -o "${output_path}" .
    if [ $? -ne 0 ]; then
        echo -e "\033[31m[编译失败] Linux ${arch}（${suffix}）架构编译出错，请检查依赖和代码！\033[0m"
        exit 1  # 编译失败则退出脚本
    fi

    # 编译成功提示
    echo -e "\033[32m[编译成功] Linux ${arch}（${suffix}）架构完成！\033[0m"
}


# -------------------------- 主流程 --------------------------
echo -e "\033[34m===== 开始执行 Go 程序交叉编译脚本 =====\033[0m"

# 1. 创建输出目录（不存在则创建，-p 确保父目录也创建）
mkdir -p "${OUTPUT_DIR}"
if [ $? -ne 0 ]; then
    echo -e "\033[31m[目录创建失败] 无法创建输出目录 ${OUTPUT_DIR}，请检查权限！\033[0m"
    exit 1
fi

# 2. 编译 Linux amd64 架构（适用于 64位 x86 Linux 系统）
build "amd64" "" "linux"

# 3. 编译 Linux armv7 架构（适用于 32位 ARM Linux 系统，如 onecloud 6.4.13）
build "arm" "7" "armv7"

build "amd64" "" "windows"
# 4. 全部编译完成
echo -e "\n\033[34m===== 所有架构编译完成！输出文件在 ${OUTPUT_DIR} 目录下 =====\033[0m"
ls -l "${OUTPUT_DIR}"  # 可选：列出输出目录文件，方便用户确认