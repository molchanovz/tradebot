package tradeplus

import (
	"log"
	"os"
)

const (
	DirectoryPath = "pkg/tradeplus/printer/"
	CodesPath     = DirectoryPath + "codes/"
	ReadyPath     = DirectoryPath + "ready/"
	GeneratedPath = DirectoryPath + "generated/"
	BatchesPath   = DirectoryPath + "batches/"
	BarcodesPath  = "assets/barcodes/"
	FontPath      = "assets/font.ttf"
)

// Progress — состояние долгой операции для сообщения в боте. Stage — что делается сейчас;
// без него бот пишет «Обработано заказов». Current из Total — счётчик этапа, при Total 0
// счётчика нет. Note — временная пометка к этапу, например о паузе перед повтором запроса.
type Progress struct {
	Stage   string
	Current int
	Total   int
	Note    string
}

// ProgressReporter шлёт этапы в канал прогресса и помнит текущий, чтобы дописать к нему Note.
// Пользоваться им можно из одной горутины; nil-репортер и nil-канал ничего не шлют.
type ProgressReporter struct {
	ch  chan Progress
	cur Progress
}

func NewProgressReporter(ch chan Progress) *ProgressReporter {
	return &ProgressReporter{ch: ch}
}

// Stage начинает этап (или обновляет его счётчик) и сбрасывает пометку.
func (r *ProgressReporter) Stage(stage string, current, total int) {
	if r == nil {
		return
	}
	r.cur = Progress{Stage: stage, Current: current, Total: total}
	r.send()
}

// Note дописывает пометку к текущему этапу.
func (r *ProgressReporter) Note(note string) {
	if r == nil {
		return
	}
	r.cur.Note = note
	r.send()
}

func (r *ProgressReporter) send() {
	if r.ch != nil {
		r.ch <- r.cur
	}
}

func CleanFiles() {
	err := os.RemoveAll(CodesPath)
	if err != nil {
		log.Println(err)
	}
	err = os.Mkdir(CodesPath, 0755)
	if err != nil {
		log.Println(err)
	}

	err = os.RemoveAll(BatchesPath)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(BatchesPath, 0755)
	if err != nil {
		log.Println(err)
	}

	err = os.RemoveAll(ReadyPath)
	if err != nil {
		log.Println(err)
	}
	err = os.Mkdir(ReadyPath, 0755)
	if err != nil {
		log.Println(err)
	}

	err = os.RemoveAll(GeneratedPath)
	if err != nil {
		log.Println(err)
	}
	err = os.Mkdir(GeneratedPath, 0755)
	if err != nil {
		log.Println(err)
	}
}

func CreateDirectories() {
	err := os.MkdirAll(GeneratedPath, 0755)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(ReadyPath, 0755)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(CodesPath, 0755)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(BatchesPath, 0755)
	if err != nil {
		log.Println(err)
	}
}
