package stickersFbs

import (
	"encoding/base64"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/jung-kurt/gofpdf"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
	"tradebot/pkg/api/wb"
	"tradebot/pkg/fbsPrinter"
	"tradebot/pkg/google"
)

type WbManager struct {
	token        string
	googleSheets google.SheetsService
}

func NewWbManager(token string) WbManager {
	return WbManager{
		token:        token,
		googleSheets: google.NewSheetsService("token.json", "credentials.json"),
	}
}

func (m WbManager) GetReadyFile(supplyId string, progressChan chan fbsPrinter.Progress) ([]string, error) {
	fbsPrinter.CreateDirectories()

	orders, err := wb.GetOrdersFbs(m.token, supplyId)
	if err != nil {
		return nil, err
	}

	totalOrders := len(orders)
	var resultFiles []string
	var ordersSlice []string
	batchCount := 0

	for i, order := range orders {
		stickers, err := wb.GetStickersFbs(m.token, order.ID)

		if err != nil {
			return nil, err
		}

		if len(stickers.Stickers) == 0 {
			continue
		}

		decodeToPDF(stickers.Stickers[0].File, stickers.Stickers[0].OrderId, order)
		ordersSlice = append(ordersSlice, fbsPrinter.ReadyPath+strconv.Itoa(order.ID)+".pdf")

		if i%10 == 0 {
			time.Sleep(2 * time.Second)
		}

		// Батчи по 300 заказов
		if (i+1)%300 == 0 || i == totalOrders-1 {
			batchCount++

			if len(ordersSlice) == 0 {
				continue // Пропускаем пустые батчи
			}

			// Создаем PDF для текущего батча
			batchFilePath := fmt.Sprintf("%s%s_%d.pdf", fbsPrinter.DirectoryPath, supplyId, batchCount)
			err = mergePDFsInDirectory(ordersSlice, batchFilePath)
			if err != nil {
				return nil, fmt.Errorf("ошибка объединения PDF для батча %d: %v", batchCount, err)
			}

			if !fileExists(batchFilePath) {
				return nil, fmt.Errorf("итоговый PDF для батча %d не создан", batchCount)
			}

			resultFiles = append(resultFiles, batchFilePath)
			ordersSlice = []string{} // Сбрасываем для следующего батча
		}

		progressChan <- fbsPrinter.Progress{Current: i + 1, Total: totalOrders}
	}

	if len(resultFiles) == 0 {
		return nil, fmt.Errorf("не было создано ни одного PDF файла")
	}

	return resultFiles, nil
}

func decodeToPNG(base64String string, orderId int) string {
	// Ваш base64 закодированный контент
	base64Data := base64String

	// Декодирование base64 в байты
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Println("Ошибка при декодировании base64:", err)
	}

	// Определите путь и имя файла для сохранения
	filePath := fbsPrinter.CodesPath + strconv.Itoa(orderId) + ".png" // Замените на желаемое имя файла и расширение

	// Открытие файла для записи
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Ошибка при создании файла:", err)
	}
	defer file.Close()

	// Запись данных в файл
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}

	return filePath
}

func decodeToPDF(base64String string, orderId int, order wb.OrderWB) {
	pageWidthMM := 75.0
	pageHeightMM := 120.0
	// Создание нового PDF-документа
	pdf := gofpdf.New("P", "mm", "", "")
	// Добавление страницы с заданными размерами
	pdf.AddPageFormat("P", gofpdf.SizeType{pageWidthMM, pageHeightMM})
	// Путь к первому PNG-файлу
	imgPath1 := decodeToPNG(base64String, orderId)
	// Добавление первого изображения в PDF (без изменения размера изображения)
	pdf.ImageOptions(imgPath1, (75-58)/2, 13, 58, 40, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	var skuImageUrl string

	skuImageUrl = fbsPrinter.BarcodesPath + order.Article + ".png"

	if !fileExists(skuImageUrl) {
		skuImageUrl = ""
	}

	if skuImageUrl == "" {
		// Путь к пустому баркоду с артикулом
		skuImageUrl = fbsPrinter.GeneratedPath + order.Article + "_generated.png"
		err := createBarcodeWithSKU(order.Article, skuImageUrl, 40)
		if err != nil {
			log.Printf("Ошибка при создании изображения с артикулом: %v", err)
			skuImageUrl = fbsPrinter.BarcodesPath + "0.png" // Резервный пустой баркод
		}
	}

	pdf.ImageOptions(skuImageUrl, (75-58)/2, 67, 58, 40, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	// Сохранение PDF-документа
	err := pdf.OutputFileAndClose(fbsPrinter.ReadyPath + strconv.Itoa(orderId) + ".pdf")
	if err != nil {
		log.Fatalf("Error saving PDF: %s", err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir() // Проверяем, что это файл, а не директория
}

func mergePDFsInDirectory(orderSlice []string, outputFile string) error {

	// Проверяем, есть ли PDF-файлы для объединения
	if len(orderSlice) == 0 {
		return fmt.Errorf("нет PDF-файлов в директории")
	}

	// Формируем команду для выполнения merge PDF через pdfcpu
	args := append([]string{"merge", outputFile}, orderSlice...)
	cmd := exec.Command("pdfcpu", args...)

	// Запуск команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка выполнения pdfcpu: %v, %s", err, string(output))
	}
	return nil
}

// Функция для создания изображения с текстом (артикул товара) и сохранения в PNG
func createBarcodeWithSKU(sku string, outputPath string, fontSize float64) error {
	const imgWidth = 580
	const imgHeight = 400

	// Создание нового изображения
	dc := gg.NewContext(imgWidth, imgHeight)

	// Установка фона (белый цвет)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Загрузка шрифта и установка его размера
	if err := dc.LoadFontFace(fbsPrinter.FontPath, fontSize); err != nil {
		return err
	}

	// Настройка текста
	dc.SetRGB(0, 0, 0) // Цвет текста (черный)
	text := sku

	// Добавление текста на изображение
	dc.DrawStringAnchored(text, float64(imgWidth)/2, float64(imgHeight)/2, 0.5, 0.5)

	// Сохранение изображения в формате PNG
	return dc.SavePNG(outputPath)
}
