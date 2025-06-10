package yandex_stickers_fbs

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
	"tradebot/pkg/fbsPrinter"

	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"strconv"
	"tradebot/pkg/api/yandex"
)

//func GetOrderInfo(token, orderId string) {
//	info, err := API.OrderInfo(token, orderId)
//	if err != nil {
//		return
//	}
//}

const (
	YaDirectoryPath = "/app/pkg/YANDEX/yandex_stickers_fbs/"

	codesPath     = YaDirectoryPath + "codes/"
	readyPath     = YaDirectoryPath + "ready/"
	generatedPath = YaDirectoryPath + "generated/"
	barcodesPath  = "/app/pkg/barcodes/"
	fontPath      = "/app/assets/font.ttf"
)

type Manager struct {
	yandexCampaignIdFBO, yandexCampaignIdFBS, token string
}

func NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token string) *Manager {
	return &Manager{
		yandexCampaignIdFBO: yandexCampaignIdFBO,
		yandexCampaignIdFBS: yandexCampaignIdFBS,
		token:               token,
	}
}

func (m Manager) GetOrdersInfo(supplyId string, progressChan chan fbsPrinter.Progress) (string, error) {
	CreateDirectories()
	orderIds, err := yandex.GetOrdersIds(m.token, supplyId)
	if err != nil {
		return "", fmt.Errorf("ошибка в GetOrdersIds: %v", err)
	}

	totalOrders := len(orderIds)

	var ordersSlice []string
	for i, orderId := range orderIds {
		//Получаем товары в заказе
		order, err := yandex.GetOrder(m.token, orderId)
		if err != nil {
			return "", fmt.Errorf("ошибка в GetOrder: %v", err)
		}
		//Получаем стикеры к товарам

		stickers, err := yandex.GetStickers(m.token, orderId)
		if err != nil {
			return "", fmt.Errorf("ошибка в GetStickers, %v", err)
		}

		// Создаем файл для записи данных
		file, err := os.Create(fmt.Sprintf("%v.pdf", codesPath+strconv.Itoa(int(order.Order.Id))))
		if err != nil {
			return "", err
		}

		// Записываем строку в файл
		_, err = file.Write([]byte(stickers))
		if err != nil {
			return "", fmt.Errorf("ошибка в записи в файл: %v", err)
		}

		file.Close()

		pdf, err := CreateLabel(fmt.Sprintf("%v.pdf", codesPath+strconv.Itoa(int(order.Order.Id))), order.Order.Items)
		if err != nil {
			return "", fmt.Errorf("ошибка в CreateLabel: %v", err)
		}

		// Сохраняем итоговый PDF
		err = pdf.OutputFileAndClose(fmt.Sprintf("%v.pdf", readyPath+strconv.Itoa(int(order.Order.Id))))
		if err != nil {
			return "", fmt.Errorf("ошибка при сохранении PDF: %v", err)
		} else {
			fmt.Println("PDF успешно создан:", fmt.Sprintf("%v.pdf", readyPath+strconv.Itoa(int(order.Order.Id))))
		}
		ordersSlice = append(ordersSlice, fmt.Sprintf("%v.pdf", readyPath+strconv.Itoa(int(order.Order.Id))))

		progressChan <- fbsPrinter.Progress{Current: i + 1, Total: totalOrders}
	}

	readyFilePath := YaDirectoryPath + supplyId + ".pdf"

	err = mergePDFsInDirectory(ordersSlice, readyFilePath)
	if err != nil {
		return "", err
	}
	if !fileExists(readyFilePath) {
		return "", fmt.Errorf("такого файла не существует")
	}

	return readyFilePath, nil
}

func CreateLabel(codePath string, items yandex.Items) (*gofpdf.Fpdf, error) {
	// Создаем новый PDF-документ
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 75, Ht: 120}, // Размер листа 75 x 120 мм
	})
	pdf.SetMargins(0, 0, 0) // Убираем отступы

	// Извлекаем страницы из первого PDF-файла
	pages, err := extractPagesFromPDF(codePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении PDF: %v", err)
	}

	var stringItems []string

	for _, item := range items {
		for _ = range item.Count {
			stringItems = append(stringItems, item.OfferId)
		}
	}

	// Добавляем каждую страницу и изображение в новый PDF
	for i, page := range pages {
		// Добавляем новую страницу
		pdf.AddPage()

		// Сохраняем страницу как временное изображение
		pageImagePath := fmt.Sprintf("page%d.jpg", i+1)
		err = saveImageToFile(page, pageImagePath)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сохранении страницы как изображения: %v", err)
		}

		// Размещаем страницу из первого файла (58 x 40 мм) в верхней части листа
		pdf.Image(pageImagePath, (75-58)/2, 10, 58, 40, false, "", 0, "") // Центрируем по горизонтали

		// Размещаем изображение из PNG-файла ниже страницы (58 x 40 мм)
		skuImageUrl := fmt.Sprintf("%v.png", barcodesPath+stringItems[i])
		isExist := fileExists(skuImageUrl)

		if !isExist {
			skuImageUrl = ""
		}

		if skuImageUrl == "" {
			// Путь к пустому баркоду с артикулом
			skuImageUrl = generatedPath + stringItems[i] + "_generated.png"
			err := createBarcodeWithSKU(stringItems[i], skuImageUrl, 40)
			if err != nil {
				log.Printf("Ошибка при создании изображения с артикулом: %v", err)
				skuImageUrl = barcodesPath + "0.png" // Резервный пустой баркод
			}
		}

		pdf.Image(skuImageUrl, (75-58)/2, 70, 58, 40, false, "", 0, "") // Центрируем по горизонтали

		// Удаляем временное изображение страницы
		err := os.Remove(pageImagePath)
		if err != nil {
			return nil, err
		}
	}
	return pdf, nil
}

// Функция для извлечения страниц из PDF
func extractPagesFromPDF(pdfPath string) ([]image.Image, error) {
	// Открываем PDF-файл
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть PDF: %v", err)
	}
	defer doc.Close()

	// Извлекаем страницы
	var pages []image.Image
	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return nil, fmt.Errorf("ошибка при извлечении страницы %d: %v", i, err)
		}
		pages = append(pages, img)
	}

	return pages, nil
}

// Функция для сохранения изображения в файл
func saveImageToFile(img image.Image, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %v", err)
	}
	defer file.Close()

	// Сохраняем изображение в формате JPEG
	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return fmt.Errorf("ошибка при сохранении изображения: %v", err)
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir() // Проверяем, что это файл, а не директория
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

func CleanFiles(supplyId string) {
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
	err = os.Remove(YaDirectoryPath + supplyId + ".pdf")
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

func createBarcodeWithSKU(sku string, outputPath string, fontSize float64) error {
	const imgWidth = 580  // Ширина изображения в пикселях
	const imgHeight = 400 // Высота изображения в пикселях

	// Создание нового изображения
	dc := gg.NewContext(imgWidth, imgHeight)

	// Установка фона (белый цвет)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Загрузка шрифта и установка его размера
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
