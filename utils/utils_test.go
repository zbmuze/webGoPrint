package utils // 与被测试函数在同一个包

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/signintech/gopdf"
)

// -------------------------- 测试 ConvertImageToPDF --------------------------
// TestConvertImageToPDF_Normal：正常场景测试（传入合法图片，验证PDF生成成功）
func TestConvertImageToPDF_Normal(t *testing.T) {
	// 1. 准备测试数据（临时文件，避免污染本地环境）
	// 1.1 创建临时测试图片（用1x1像素的空白图片模拟，也可使用真实图片路径）
	tempImgDir := t.TempDir() // 自动清理的临时目录
	tempImgPath := filepath.Join(tempImgDir, "test.jpg")
	createTempImage(tempImgPath, t) // 辅助函数：生成临时图片

	// 1.2 定义输出PDF路径（临时文件）
	tempPdfPath := filepath.Join(tempImgDir, "output.pdf")

	// 2. 调用被测试函数
	err := ConvertImageToPDF(tempImgPath, tempPdfPath)
	if err != nil {
		t.Fatalf("正常场景测试失败：调用ConvertImageToPDF返回错误 = %v", err)
	}

	// 3. 验证结果（检查PDF是否生成、是否合法）
	// 3.1 检查PDF文件是否存在
	if _, err := os.Stat(tempPdfPath); os.IsNotExist(err) {
		t.Fatalf("正常场景测试失败：PDF文件未生成，路径 = %s", tempPdfPath)
	}

	// 3.2 检查PDF是否合法（用gopdf打开验证）
	pdf := gopdf.GoPdf{}
	if err := pdf.OpenFile(tempPdfPath); err != nil {
		t.Fatalf("正常场景测试失败：生成的PDF不合法，打开错误 = %v", err)
	}
	defer pdf.CloseFile()

	// 3.3 检查PDF页数（正常应为1页）
	pageCount, err := pdf.GetPageCount()
	if err != nil || pageCount != 1 {
		t.Fatalf("正常场景测试失败：PDF页数异常，实际页数 = %d，错误 = %v", pageCount, err)
	}

	t.Log("正常场景测试通过！")
}

// TestConvertImageToPDF_Error：错误场景测试（传入不存在的图片，验证返回错误）
func TestConvertImageToPDF_Error(t *testing.T) {
	// 1. 准备测试数据（不存在的图片路径）
	nonExistentImgPath := "non_existent.jpg"
	tempPdfPath := filepath.Join(t.TempDir(), "output.pdf")

	// 2. 调用被测试函数（预期会返回错误）
	err := ConvertImageToPDF(nonExistentImgPath, tempPdfPath)
	if err == nil {
		t.Fatalf("错误场景测试失败：传入不存在的图片，却未返回错误")
	}

	// 3. 验证结果（PDF不应生成）
	if _, err := os.Stat(tempPdfPath); !os.IsNotExist(err) {
		t.Fatalf("错误场景测试失败：传入错误图片却生成了PDF，路径 = %s", tempPdfPath)
	}

	t.Logf("错误场景测试通过！错误信息符合预期：%v", err)
}
