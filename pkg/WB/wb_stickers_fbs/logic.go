package wb_stickers_fbs

import (
	"encoding/base64"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/jung-kurt/gofpdf"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	DirectoryPath = "pkg/WB/wb_stickers_fbs/"

	codesPath     = DirectoryPath + "codes/"
	barcodesPath  = DirectoryPath + "barcodes/"
	generatedPath = DirectoryPath + "generated/"
	readyPath     = DirectoryPath + "ready/"
)

func GetReadyFile(wildberriesKey, supplyId string) error {
	orders, err := GetOrdersFbs(wildberriesKey, supplyId)
	if err != nil {
		return err
	}
	var ordersSlice []string
	for _, order := range orders {
		stickers := GetStickersFbs(wildberriesKey, order.ID)
		decodeToPDF(stickers.Stickers[0].File, stickers.Stickers[0].OrderId, order)
		ordersSlice = append(ordersSlice, readyPath+strconv.Itoa(order.ID)+".pdf")
	}
	err = mergePDFsInDirectory(ordersSlice, DirectoryPath+supplyId+".pdf")
	if err != nil {
		return err
	}
	if !fileExists(DirectoryPath + supplyId + ".pdf") {
		err = fmt.Errorf("такого файла не существует")
	}
	return err
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
	filePath := codesPath + strconv.Itoa(orderId) + ".png" // Замените на желаемое имя файла и расширение

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

func decodeToPDF(base64String string, orderId int, order OrderWB) {
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
	for _, barcode := range order.SKUs {
		skuImageUrl = barcodesPath + barcode + ".png"
		if fileExists(skuImageUrl) {
			break
		}
		skuImageUrl = ""
	}

	if skuImageUrl == "" {
		// Путь к пустому баркоду с артикулом
		skuImageUrl = generatedPath + order.Article + "_generated.png"
		err := createBarcodeWithSKU(order.Article, skuImageUrl, 40)
		if err != nil {
			log.Printf("Ошибка при создании изображения с артикулом: %v", err)
			skuImageUrl = barcodesPath + "0.png" // Резервный пустой баркод
		}
	}

	pdf.ImageOptions(skuImageUrl, (75-58)/2, 67, 58, 40, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	// Сохранение PDF-документа
	err := pdf.OutputFileAndClose(readyPath + strconv.Itoa(orderId) + ".pdf")
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

func Clean_files(supplyId string) {
	err := os.RemoveAll(codesPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(readyPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(codesPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(generatedPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(generatedPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(readyPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove(DirectoryPath + supplyId + ".pdf")
	if err != nil {
		fmt.Println(err)
	}

}

// Функция для создания изображения с текстом (артикул товара) и сохранения в PNG
func createBarcodeWithSKU(sku string, outputPath string, fontSize float64) error {
	const imgWidth = 580  // Ширина изображения в пикселях
	const imgHeight = 400 // Высота изображения в пикселях

	// Создание нового изображения
	dc := gg.NewContext(imgWidth, imgHeight)

	// Установка фона (белый цвет)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Загрузка шрифта и установка его размера
	fontPath := "font.ttf" // Укажите путь к вашему TTF-шрифту
	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
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
