package utils

import (
	"image"
	_ "image/jpeg" // 支持JPG图片格式
	_ "image/png"  // 支持PNG图片格式
	"os"

	"github.com/signintech/gopdf"
)

// ConvertImageToPDF 将图片转换为A4尺寸的PDF，保持图片比例并居中显示
//
//	imagePath : 输入图片地址
//	outputPath : 保存PDF地址
func ConvertImageToPDF(imagePath, outputPath string) error {
	// 创建PDF实例
	pdf := gopdf.GoPdf{}

	// 初始化PDF，设置为A4页面尺寸
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4, // 使用内置A4配置
	})

	// 添加一个新页面
	pdf.AddPage()

	// 打开图片文件
	imgFile, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer imgFile.Close()

	// 解码图片获取原始尺寸
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	// 获取图片原始宽高（像素）
	imgWidth := float64(img.Bounds().Dx())
	imgHeight := float64(img.Bounds().Dy())

	// 获取A4页面尺寸（单位：点，1点=1/72英寸）
	pageWidth := gopdf.PageSizeA4.W
	pageHeight := gopdf.PageSizeA4.H

	// 定义页边距
	margin := 20.0

	// 计算内容区域尺寸（减去边距）
	contentWidth := pageWidth - (2 * margin)
	contentHeight := pageHeight - (2 * margin)

	// 计算宽度和高度地缩放比例
	widthRatio := contentWidth / imgWidth
	heightRatio := contentHeight / imgHeight

	// 选择较小的缩放比例以确保图片完全适应页面
	scaleRatio := widthRatio
	if heightRatio < scaleRatio {
		scaleRatio = heightRatio
	}

	// 计算缩放后的图片尺寸
	scaledWidth := imgWidth * scaleRatio
	scaledHeight := imgHeight * scaleRatio

	// 计算居中位置
	posX := (pageWidth - scaledWidth) / 2
	posY := (pageHeight - scaledHeight) / 2

	// 将图片添加到PDF指定位置
	if err := pdf.Image(imagePath, posX, posY, &gopdf.Rect{
		W: scaledWidth,
		H: scaledHeight,
	}); err != nil {
		return err
	}

	// 保存PDF到文件
	return pdf.WritePdf(outputPath)
}

// MergePDFs 将多个PDF文件合并为一个，按 inputPaths 顺序排列页面
// inputPaths: 待合并的PDF文件路径切片（如 []string{"a.pdf", "b.pdf"}）
// outputPath: 合并后的PDF输出路径（如 "merged.pdf"）
// return: 合并过程中的错误（如文件不存在、格式错误等）
//func MergePDFs(inputPaths []string, outputPath string) error {
//	// 创建PDF实例
//	pdf := gopdf.GoPdf{}
//
//	// 初始化PDF，设置为A4页面尺寸
//	pdf.Start(gopdf.Config{
//		PageSize: *gopdf.PageSizeA4, // 使用内置A4配置
//	})
//
//	// 初始化PDF导入器
//	importer := gofpdi.NewImporter()
//	// 定义每张纸放置的页面数量（2x2 网格）
//	pagesPerSheet := 4
//	// 计算每个小页面的宽度和高度（A4纸平分4份，考虑边距）
//	margin := 20.0
//	usableWidth := 595.28 - (2 * margin)
//	usableHeight := 841.89 - (2 * margin)
//	cellWidth := usableWidth / 2
//	cellHeight := usableHeight / 2
//
//	currentPage := gopdf.GoPdf{} // 临时变量，实际需调整逻辑
//	pdf.AddPage()                // 添加第一张纸
//
//	for index, pdfFile := range inputPaths {
//		// 计算当前页面在当前纸张上的位置（行和列）
//		positionOnPage := index % pagesPerSheet
//		col := float64(positionOnPage % 2)
//		row := float64(positionOnPage / 2)
//
//		// 计算当前小页面的左上角坐标
//		x := margin + (col * cellWidth)
//		y := margin + (row * cellHeight)
//
//		// 导入源PDF页面:cite[1]
//		importer.SetSourceFile(pdfFile)
//		// 假设每个文件都是单页PDF
//		importedPage := importer.ImportPage(1, "/MediaBox")
//
//		// 将导入的页面缩放并放置到当前纸张的指定位置:cite[1]
//		// 这里假设缩放至小单元格的大小
//		importer.UseImportedTemplate(pdf, importedPage, x, y, cellWidth, cellHeight)
//
//		// 如果当前纸张已放满，且还有后续页面，则添加新纸张
//		if (index+1)%pagesPerSheet == 0 && (index+1) < len(pdfFiles) {
//			pdf.AddPage()
//		}
//	}
//
//	err := pdf.WritePdf("nup_merged.pdf")
//	if err != nil {
//		panic(err)
//	}
//	return pdf.WritePdf(outputPath)
//}
