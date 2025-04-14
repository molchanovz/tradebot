package stickersFBS

import (
	"WildberriesGo_bot/pkg/api/ozon"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	OzonDirectoryPath = "pkg/OZON/stickersFBS/"
	codesPath         = OzonDirectoryPath + "codes/"
	readyPath         = OzonDirectoryPath + "ready/"
	generatedPath     = OzonDirectoryPath + "generated/"
	barcodesPath      = "pkg/barcodes/"
)

type OzonManager struct {
	clientId, token string
}

func NewOzonManager(clientId, token string) OzonManager {
	return OzonManager{
		clientId: clientId,
		token:    token,
	}
}

func (m OzonManager) GetLabels() error {
	since := time.Now().AddDate(0, 0, -7).Format("2006-01-02T15:04:05.000Z")
	to := time.Now().Format("2006-01-02T15:04:05.000Z")

	orderIds := ozon.PostingsListFbs(m.clientId, m.token, since, to, 0, "awaiting_deliver")

	var ordersSlice []string
	for _, order := range orderIds.Result.PostingsFBS {

		//Получаем стикеры к товарам
		label := ozon.V2PostingFbsPackageLabel(m.clientId, m.token, order.PostingNumber)

		// Создаем файл для записи данных
		file, err := os.Create(fmt.Sprintf("%v.pdf", codesPath+order.PostingNumber))
		if err != nil {
			return err
		}

		// Записываем строку в файл
		_, err = file.Write([]byte(label))
		if err != nil {
			return fmt.Errorf("ошибка в записи в файл: %v", err)
		}

		file.Close()

		ordersSlice = append(ordersSlice, fmt.Sprintf("%v.pdf", codesPath+order.PostingNumber))
	}

	err := mergePDFsInDirectory(ordersSlice, OzonDirectoryPath+"ozon.pdf")
	if err != nil {
		return err
	}
	if !fileExists(OzonDirectoryPath + "ozon.pdf") {
		return fmt.Errorf("такого файла не существует")
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
