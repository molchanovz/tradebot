package fbsPrinter

import (
	"fmt"
	"os"
)

const (
	DirectoryPath = "/app/pkg/fbsPrinter/"
	CodesPath     = DirectoryPath + "codes/"
	ReadyPath     = DirectoryPath + "ready/"
	GeneratedPath = DirectoryPath + "generated/"
	BatchesPath   = DirectoryPath + "batches/"
	BarcodesPath  = "/assets/barcodes/"
	FontPath      = "/assets/font.ttf"
)

type Progress struct {
	Current int
	Total   int
}

func CleanFiles() {
	err := os.RemoveAll(CodesPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(CodesPath, 0755)
	if err != nil {
		fmt.Println(err)
	}

	err = os.RemoveAll(BatchesPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(BatchesPath, 0755)
	if err != nil {
		fmt.Println(err)
	}

	err = os.RemoveAll(ReadyPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(ReadyPath, 0755)
	if err != nil {
		fmt.Println(err)
	}

	err = os.RemoveAll(GeneratedPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(GeneratedPath, 0755)
	if err != nil {
		fmt.Println(err)
	}

}

func CreateDirectories() {
	err := os.MkdirAll(GeneratedPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(ReadyPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(CodesPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir(BatchesPath, 0755)
	if err != nil {
		fmt.Println(err)
	}
}
