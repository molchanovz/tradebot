package ozon

import (
	"errors"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tradebot/pkg/client/ozon"
	"tradebot/pkg/tradeplus"

	"github.com/fogleman/gg"
	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type StickerManager struct {
	clientID, token string
	printedOrders   map[string]struct{}
}

func NewStickerManager(clientID, token string, printedOrders map[string]struct{}) StickerManager {
	return StickerManager{
		clientID:      clientID,
		token:         token,
		printedOrders: printedOrders,
	}
}

const (
	AllLabels = "all"
	NewLabels = "new"
)

func (m StickerManager) GetAllLabels(progressChan chan tradeplus.Progress) ([]string, error) {
	tradeplus.CreateDirectories()

	orderIds, err := m.getSortedFbsOrders()
	if err != nil {
		return []string{}, err
	}

	readyPdfPath, err := m.getReadyPdf(orderIds, progressChan)
	if err != nil {
		log.Println("Ошибка при получении файла:", err)
	}

	return readyPdfPath, nil
}

func (m StickerManager) GetNewLabels(progressChan chan tradeplus.Progress) ([]string, ozon.PostingslistFbs, error) {
	tradeplus.CreateDirectories()

	orders, err := m.getSortedFbsOrders()
	if err != nil {
		return []string{}, ozon.PostingslistFbs{}, err
	}

	//Проверка, есть ли новые заказы
	newOrders := ozon.PostingslistFbs{}
	for _, p := range orders.Result.PostingsFBS {
		if _, ok := m.printedOrders[p.PostingNumber]; !ok {
			newOrders.Result.PostingsFBS = append(newOrders.Result.PostingsFBS, p)
		}
	}

	if len(newOrders.Result.PostingsFBS) == 0 {
		return []string{}, newOrders, ErrNoRows
	}

	readyPdfPaths, err := m.getReadyPdf(newOrders, progressChan)
	if err != nil {
		return []string{}, newOrders, err
	}

	return readyPdfPaths, newOrders, nil
}

func (m StickerManager) getReadyPdf(orderIds ozon.PostingslistFbs, progressChan chan tradeplus.Progress) ([]string, error) {
	tradeplus.CreateDirectories()

	totalOrders := len(orderIds.Result.PostingsFBS)
	var resultFiles []string
	var combinedPDFs []string
	batchCount := 0

	for i, order := range orderIds.Result.PostingsFBS {
		// 1. Скачиваем этикетку Ozon
		labelPDF, err := ozon.NewClient(m.clientID, m.token).Labels(order.PostingNumber)
		if err != nil {
			return nil, fmt.Errorf("get labels failed: %w", err)
		}

		if labelPDF == "" {
			return nil, fmt.Errorf("label %s is nil", order.PostingNumber)
		}

		// 2. Сохраняем во временный файл
		orderPDF := fmt.Sprintf("%v.pdf", tradeplus.CodesPath+order.PostingNumber)
		if err = os.WriteFile(orderPDF, []byte(labelPDF), 0644); err != nil {
			return nil, fmt.Errorf("ошибка записи PDF: %w", err)
		}

		// 3. Извлекаем первую страницу
		if err = extractFirstPage(orderPDF); err != nil {
			return nil, fmt.Errorf("ошибка извлечения страницы: %w", err)
		}

		// 4. Объединяем с баркодом
		finalPDF := fmt.Sprintf("%v.pdf", tradeplus.ReadyPath+order.PostingNumber)
		if err = combineLabelWithBarcode(orderPDF, finalPDF, order.Products[0].OfferID); err != nil {
			return nil, fmt.Errorf("ошибка объединения PDF с баркодом: %w", err)
		}

		// 5. Удаляем временные файлы
		os.Remove(orderPDF)

		combinedPDFs = append(combinedPDFs, finalPDF)

		// Батчи по 200 заказов
		if (i+1)%200 == 0 || i == len(orderIds.Result.PostingsFBS)-1 {
			batchCount++

			if len(combinedPDFs) == 0 {
				return nil, fmt.Errorf("нет объединенных файлов для батча %d", batchCount)
			}

			// 6. Объединяем PDF текущего батча
			readyPdfPath := fmt.Sprintf("%s/ozon_%d.pdf", tradeplus.BatchesPath, batchCount)
			if err = mergePDFsInDirectory(combinedPDFs, readyPdfPath); err != nil {
				return nil, fmt.Errorf("ошибка объединения PDF для батча %d: %v", batchCount, err)
			}

			// 7. Проверяем результат
			if !fileExists(readyPdfPath) {
				return nil, fmt.Errorf("итоговый PDF для батча %d не создан", batchCount)
			}

			resultFiles = append(resultFiles, readyPdfPath)
			combinedPDFs = []string{}
		}

		if i%5 == 0 {
			progressChan <- tradeplus.Progress{Current: i, Total: totalOrders}
		}
	}

	if len(resultFiles) == 0 {
		return nil, errors.New("не было создано ни одного PDF файла")
	}

	return resultFiles, nil
}

func (m StickerManager) getSortedFbsOrders() (ozon.PostingslistFbs, error) {
	since := time.Now().AddDate(0, 0, -7).Format("2006-01-02T15:04:05.000Z")
	to := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05.000Z")

	// Обработка FBS заказов
	var orders ozon.PostingslistFbs
	offset := 0
	limit := 1000
	for {
		postingsListFbs, err := ozon.NewClient(m.clientID, m.token).PostingsListFbs(since, to, offset, "awaiting_deliver")
		if err != nil {
			return postingsListFbs, err
		}

		orders.Result.PostingsFBS = append(orders.Result.PostingsFBS, postingsListFbs.Result.PostingsFBS...)

		if !postingsListFbs.Result.HasNext || len(postingsListFbs.Result.PostingsFBS) < limit {
			break
		}
		offset += len(postingsListFbs.Result.PostingsFBS)
	}

	if len(orders.Result.PostingsFBS) == 0 {
		return orders, ErrNoRows
	}

	sort.Slice(orders.Result.PostingsFBS, func(i, j int) bool {
		return orders.Result.PostingsFBS[i].Products[0].OfferID < orders.Result.PostingsFBS[j].Products[0].OfferID
	})
	return orders, nil
}

func combineLabelWithBarcode(ozonPdfPath, outputPath, article string) error {
	tmpImg := ozonPdfPath + ".jpg"

	defer os.Remove(tmpImg)

	doc, err := fitz.New(ozonPdfPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия PDF: %w", err)
	}
	defer doc.Close()

	img, err := doc.Image(0)
	if err != nil {
		return fmt.Errorf("ошибка извлечения страницы: %w", err)
	}

	file, err := os.Create(tmpImg)
	if err != nil {
		return fmt.Errorf("ошибка создания JPEG: %w", err)
	}
	defer file.Close()

	if err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("ошибка сохранения JPEG: %w", err)
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
	barcodePath := tradeplus.BarcodesPath + article + ".png"
	if !fileExists(barcodePath) {
		barcodePath = tradeplus.GeneratedPath + article + "_generated.png"
		if err = createBarcodeWithSKU(article, barcodePath, 40); err != nil {
			log.Println(err)
			barcodePath = tradeplus.BarcodesPath + "0.png"
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
	if err := dc.LoadFontFace(tradeplus.FontPath, fontSize); err != nil {
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
		return fmt.Errorf("ошибка при создании временной директории: %w", err)
	}
	defer os.RemoveAll(tempDir) // Удаляем временную директорию в конце

	// Выбираем только первую страницу
	selectedPages := []string{"1"}

	// Извлекаем первую страницу во временную директорию
	if err := api.ExtractPagesFile(pdfPath, tempDir, selectedPages, nil); err != nil {
		return fmt.Errorf("ошибка при извлечении страницы: %w", err)
	}

	// Получаем путь к извлеченному файлу (ожидаем формат filename_page_1.pdf)
	base := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
	extractedFile := filepath.Join(tempDir, base+"_page_1.pdf")

	// Удаляем оригинальный файл
	if err := os.Remove(pdfPath); err != nil {
		return fmt.Errorf("ошибка при удалении оригинального файла: %w", err)
	}

	// Перемещаем извлеченный файл на место оригинального
	if err := os.Rename(extractedFile, pdfPath); err != nil {
		return fmt.Errorf("ошибка при перемещении файла: %w", err)
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

func mergePDFsInDirectory(orderSlice []string, outputFile string) error {
	// Проверяем, есть ли PDF-файлы для объединения
	if len(orderSlice) == 0 {
		return errors.New("нет PDF-файлов в директории")
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
