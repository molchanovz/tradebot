package stickersFBS

import (
	"errors"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tradebot/pkg/api/ozon"
)

const (
	OzonDirectoryPath = "app/pkg/OZON/stickersFBS/"
	codesPath         = OzonDirectoryPath + "codes/"
	readyPath         = OzonDirectoryPath + "ready/"
	generatedPath     = OzonDirectoryPath + "generated/"
	barcodesPath      = "/assets/barcodes/"
)

//const (
//	OzonDirectoryPath = "pkg/marketplaces/OZON/stickersFBS/"
//	codesPath         = OzonDirectoryPath + "codes/"
//	readyPath         = OzonDirectoryPath + "ready/"
//	generatedPath     = OzonDirectoryPath + "generated/"
//	barcodesPath      = "assets/barcodes/"
//)

type OzonManager struct {
	clientId, token string
	printedOrders   map[string]struct{}
}

func NewOzonManager(clientId, token string, printedOrders map[string]struct{}) OzonManager {
	return OzonManager{
		clientId:      clientId,
		token:         token,
		printedOrders: printedOrders,
	}
}

const (
	AllLabels = "all"
	NewLabels = "new"
)

func (m OzonManager) GetAllLabels() (string, error) {
	orderIds := m.getSortedFbsOrders()

	readyPdfPath, err := m.getReadyPdf(orderIds)
	if err != nil {
		return "", err
	}

	return readyPdfPath, nil
}

func (m OzonManager) GetNewLabels() (string, ozon.PostingslistFbs, error) {
	orders := m.getSortedFbsOrders()

	//Проверка, есть ли новые заказы
	newOrders := ozon.PostingslistFbs{}
	for _, posting := range orders.Result.PostingsFBS {
		if _, ok := m.printedOrders[posting.PostingNumber]; !ok {
			newOrders.Result.PostingsFBS = append(newOrders.Result.PostingsFBS, posting)
		}
	}

	if len(newOrders.Result.PostingsFBS) == 0 {
		return "", newOrders, errors.New("Новых заказов нет")
	}

	readyPdfPath, err := m.getReadyPdf(newOrders)
	if err != nil {
		return "", newOrders, err
	}

	return readyPdfPath, newOrders, nil
}

func (m OzonManager) getReadyPdf(orderIds ozon.PostingslistFbs) (string, error) {
	CreateDirectories()

	var combinedPDFs []string
	for _, order := range orderIds.Result.PostingsFBS {
		// 1. Скачиваем этикетку Ozon
		labelPDF := ozon.V2PostingFbsPackageLabel(m.clientId, m.token, order.PostingNumber)

		// 2. Сохраняем во временный файл
		orderPDF := fmt.Sprintf("%v.pdf", codesPath+order.PostingNumber)
		if err := os.WriteFile(orderPDF, []byte(labelPDF), 0644); err != nil {
			return "", fmt.Errorf("ошибка записи PDF: %v", err)
		}

		// 3. Извлекаем первую страницу
		if err := extractFirstPage(orderPDF); err != nil {
			return "", fmt.Errorf("ошибка извлечения страницы: %v", err)
		}

		// 4. Объединяем с баркодом
		finalPDF := fmt.Sprintf("%v.pdf", readyPath+order.PostingNumber)
		if err := combineLabelWithBarcode(orderPDF, finalPDF, order.Products[0].OfferId); err != nil {
			return "", fmt.Errorf("ошибка объединения PDF с баркодом: %v", err)
		}

		// 5. Удаляем временные файлы
		os.Remove(orderPDF)

		combinedPDFs = append(combinedPDFs, finalPDF)
	}

	// 6. Объединяем все PDF в один
	readyPdfPath := OzonDirectoryPath + "ozon.pdf"
	if err := mergePDFsInDirectory(combinedPDFs, readyPdfPath); err != nil {
		return "", fmt.Errorf("ошибка объединения PDF: %v", err)
	}

	// 7. Проверяем результат
	if !fileExists(readyPdfPath) {
		return "", fmt.Errorf("итоговый PDF не создан")
	}
	return readyPdfPath, nil
}

func (m OzonManager) getSortedFbsOrders() ozon.PostingslistFbs {
	since := time.Now().AddDate(0, 0, -7).Format("2006-01-02T15:04:05.000Z")
	to := time.Now().Format("2006-01-02T15:04:05.000Z")

	orderIds, _ := ozon.PostingsListFbs(m.clientId, m.token, since, to, 0, "awaiting_deliver")

	sort.Slice(orderIds.Result.PostingsFBS, func(i, j int) bool {
		return orderIds.Result.PostingsFBS[i].Products[0].OfferId < orderIds.Result.PostingsFBS[j].Products[0].OfferId
	})
	return orderIds
}

func combineLabelWithBarcode(ozonPdfPath, outputPath, article string) error {
	tmpImg := ozonPdfPath + ".jpg"
	defer os.Remove(tmpImg)

	doc, err := fitz.New(ozonPdfPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия PDF: %v", err)
	}
	defer doc.Close()

	img, err := doc.Image(0)
	if err != nil {
		return fmt.Errorf("ошибка извлечения страницы: %v", err)
	}

	file, err := os.Create(tmpImg)
	if err != nil {
		return fmt.Errorf("ошибка создания JPEG: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("ошибка сохранения JPEG: %v", err)
	}

	pdf := gofpdf.New("P", "mm", "", "")
	pdf.AddPageFormat("P", gofpdf.SizeType{Wd: 75, Ht: 120})

	// Размеры изображения до поворота
	origWidth := 58.0
	origHeight := 40.0

	// После поворота ширина/высота меняются местами
	rotatedHeight := origWidth

	// Центр страницы
	pageWidth := 75.0

	// Центр изображения после поворота
	centerX := pageWidth / 2
	centerY := 13 + rotatedHeight/2 // Сдвиг вверх

	// Вставка и поворот изображения
	pdf.TransformBegin()
	pdf.TransformRotate(90, centerX, centerY)
	pdf.ImageOptions(tmpImg, centerX-origHeight/2+2, 8, origHeight+10, origWidth+20, false,
		gofpdf.ImageOptions{ImageType: "JPG"}, 0, "")
	pdf.TransformEnd()

	// Вставка баркода
	barcodePath := barcodesPath + article + ".png"
	if !fileExists(barcodePath) {
		barcodePath = generatedPath + article + "_generated.png"
		if err := createBarcodeWithSKU(article, barcodePath, 40); err != nil {
			barcodePath = barcodesPath + "0.png"
		}
	}

	// Центрируем баркод
	barcodeWidth := 58.0
	barcodeHeight := 40.0
	barcodeX := (pageWidth - barcodeWidth) / 2
	barcodeY := 67.0

	pdf.ImageOptions(barcodePath, barcodeX, barcodeY, barcodeWidth, barcodeHeight, false,
		gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	return pdf.OutputFileAndClose(outputPath)
}

func createBarcodeWithSKU(sku string, outputPath string, fontSize float64) error {
	const imgWidth = 580  // Ширина изображения в пикселях
	const imgHeight = 400 // Высота изображения в пикселях

	// Создание нового изображения
	dc := gg.NewContext(imgWidth, imgHeight)

	// Установка фона (белый цвет)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Загрузка шрифта и установка его размера
	fontPath := "/assets/font.ttf" // Укажите путь к вашему TTF-шрифту
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

// extractFirstPage извлекает первую страницу из PDF файла
func extractFirstPage(pdfPath string) error {
	// Создаём временную директорию для извлечения
	dir := filepath.Dir(pdfPath)
	tempDir := filepath.Join(dir, "_temp_pages")

	// Создаём временную директорию
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("ошибка при создании временной директории: %v", err)
	}
	defer os.RemoveAll(tempDir) // Удаляем временную директорию в конце

	// Выбираем только первую страницу
	selectedPages := []string{"1"}

	// Извлекаем первую страницу во временную директорию
	if err := api.ExtractPagesFile(pdfPath, tempDir, selectedPages, nil); err != nil {
		return fmt.Errorf("ошибка при извлечении страницы: %v", err)
	}

	// Получаем путь к извлеченному файлу (ожидаем формат filename_page_1.pdf)
	base := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
	extractedFile := filepath.Join(tempDir, base+"_page_1.pdf")

	// Удаляем оригинальный файл
	if err := os.Remove(pdfPath); err != nil {
		return fmt.Errorf("ошибка при удалении оригинального файла: %v", err)
	}

	// Перемещаем извлеченный файл на место оригинального
	if err := os.Rename(extractedFile, pdfPath); err != nil {
		return fmt.Errorf("ошибка при перемещении файла: %v", err)
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (m OzonManager) CleanFiles(supplyId string) {
	err := os.RemoveAll(codesPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(readyPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(codesPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(generatedPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(generatedPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(readyPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove(OzonDirectoryPath + supplyId + ".pdf")
	if err != nil {
		fmt.Println(err)
	}

}

func CreateDirectories() {
	err := os.MkdirAll(generatedPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(readyPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(codesPath, 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
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
