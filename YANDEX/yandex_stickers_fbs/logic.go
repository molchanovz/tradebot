package yandex_stickers_fbs

import (
	"WildberriesGo_bot/YANDEX/API"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
)

//func GetOrderInfo(token, orderId string) {
//	info, err := API.OrderInfo(token, orderId)
//	if err != nil {
//		return
//	}
//}

func GetOrdersInfo(token, supplyId string) error {
	orderIds, err := GetOrdersIds(token, supplyId)
	if err != nil {
		return fmt.Errorf("пизда в GetOrdersIds: %v", err)
	}

	var ordersSlice []string
	for _, orderId := range orderIds {
		//Получаем товары в заказе
		order, err := GetOrder(token, orderId)

		if err != nil {
			return fmt.Errorf("пизда в GetOrder: %v", err)
		}
		//Получаем стикеры к товарам

		stickers, err := API.GetStickers(token, orderId)
		if err != nil {
			return fmt.Errorf("пизда в GetStickers, %v", err)
		}

		// Создаем файл для записи данных
		file, err := os.Create(fmt.Sprintf("YANDEX/yandex_stickers_fbs/codes/%v.pdf", order.Order.Id))
		if err != nil {
			return err
		}

		// Записываем строку в файл
		_, err = file.Write([]byte(stickers))
		if err != nil {
			panic(err)
		}

		file.Close()

		pdf, err := CreateLabel(fmt.Sprintf("YANDEX/yandex_stickers_fbs/codes/%v.pdf", order.Order.Id), order.Order.Items)
		if err != nil {
			return err
		}

		// Сохраняем итоговый PDF
		err = pdf.OutputFileAndClose(fmt.Sprintf("YANDEX/yandex_stickers_fbs/ready/%v.pdf", order.Order.Id))
		if err != nil {
			fmt.Println("Ошибка при сохранении PDF:", err)
		} else {
			fmt.Println("PDF успешно создан:", fmt.Sprintf("YANDEX/yandex_stickers_fbs/ready/%v.pdf", order.Order.Id))
		}

		ordersSlice = append(ordersSlice, fmt.Sprintf("YANDEX/yandex_stickers_fbs/ready/%v.pdf", order.Order.Id))
	}

	err = mergePDFsInDirectory(ordersSlice, "YANDEX/yandex_stickers_fbs/"+supplyId+".pdf")
	if err != nil {
		return err
	}
	if !fileExists("YANDEX/yandex_stickers_fbs/" + supplyId + ".pdf") {
		return fmt.Errorf("такого файла не существует")
	}

	return nil
}

func CreateLabel(codePath string, items Items) (*gofpdf.Fpdf, error) {
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
		skuImageUrl := fmt.Sprintf("YANDEX/yandex_stickers_fbs/barcodes/%v.png", items[i].ShopSku)
		isExist := fileExists(skuImageUrl)

		if !isExist {
			skuImageUrl = ""
		}

		if skuImageUrl == "" {
			// Путь к пустому баркоду с артикулом
			skuImageUrl = "YANDEX/yandex_stickers_fbs/generated/" + items[i].ShopSku + "_generated.png"
			err := createBarcodeWithSKU(items[i].ShopSku, skuImageUrl, 40)
			if err != nil {
				log.Printf("Ошибка при создании изображения с артикулом: %v", err)
				skuImageUrl = "YANDEX/yandex_stickers_fbs/barcodes/0.png" // Резервный пустой баркод
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

func Clean_files(supplyId string) {
	err := os.RemoveAll("YANDEX/yandex_stickers_fbs/codes")
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll("YANDEX/yandex_stickers_fbs/ready")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir("YANDEX/yandex_stickers_fbs/codes", 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir("YANDEX/yandex_stickers_fbs/ready", 0755) // 0755 - это права доступа к директории (чтение, запись, выполнение)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove("YANDEX/yandex_stickers_fbs/" + supplyId + ".pdf")
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
