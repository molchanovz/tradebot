package bot

import (
	"context"
	"errors"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"log"
	"math"
	"os"
	"sync"
	"tradebot/pkg/db"
	"tradebot/pkg/fbsPrinter"
	"tradebot/pkg/marketplaces/YANDEX"
)

const (
	CallbackStartHandler                 = "start"
	MessageStartHandler                  = "/start"
	CallbackSelectOzonCabinetHandler     = "CABINET-OZON_"
	CallbackSettingsSelectCabinetHandler = "SETTINGS-CABINET_"
)

type Manager struct {
	b             *botlib.Bot
	repo          *db.Repo
	yandexService YANDEX.Service
	myChatId      string
	SheetMap      *sync.Map
	ApiMap        *sync.Map
}

func NewBotManager(yandexService YANDEX.Service, repo *db.Repo, myChatId string) *Manager {
	return &Manager{
		yandexService: yandexService,
		repo:          repo,
		myChatId:      myChatId,
		SheetMap:      new(sync.Map),
		ApiMap:        new(sync.Map),
	}
}

func (m *Manager) SetBot(bot *botlib.Bot) {
	m.b = bot
}
func (m *Manager) GetBot() *botlib.Bot {
	return m.b
}

func (m *Manager) RegisterBotHandlers() {
	m.b.RegisterHandler(botlib.HandlerTypeMessageText, MessageStartHandler, botlib.MatchTypePrefix, m.startHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackStartHandler, botlib.MatchTypePrefix, m.startHandler)
	m.b.RegisterHandler(botlib.HandlerTypeMessageText, MessageSettingsHandler, botlib.MatchTypePrefix, m.settingsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, MessageSettingsHandler, botlib.MatchTypePrefix, m.settingsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackSettingsHandler, botlib.MatchTypePrefix, m.selectMpSettingsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackSettingsSelectCabinetHandler, botlib.MatchTypePrefix, m.settingsMPHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackChangeAPIHandler, botlib.MatchTypePrefix, m.ChangeApiHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackChangeSheetHandler, botlib.MatchTypePrefix, m.ChangeSheetHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbHandler, botlib.MatchTypeExact, wbHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexHandler, botlib.MatchTypeExact, m.yandexHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonHandler, botlib.MatchTypeExact, m.ozonHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbFbsHandler, botlib.MatchTypeExact, m.stickersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexFbsHandler, botlib.MatchTypeExact, m.yandexFbsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbOrdersHandler, botlib.MatchTypePrefix, m.wbOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexOrdersHandler, botlib.MatchTypePrefix, m.yandexOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonOrdersHandler, botlib.MatchTypePrefix, m.ozonOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStocksHandler, botlib.MatchTypePrefix, m.ozonStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbStocksHandler, botlib.MatchTypePrefix, m.wbStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStickersHandler, botlib.MatchTypePrefix, m.ozonStickersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonPrintStickersHandler, botlib.MatchTypePrefix, m.ozonPrintStickers)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackSelectOzonCabinetHandler, botlib.MatchTypePrefix, m.ozonCabinetHandler)

	//b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypePrefix, wbOrdersHandler)

}

// DefaultHandler ловит сообщения без команд, проверяет статус пользователя, после обновляет статус на enabled
func (m *Manager) DefaultHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	message := update.Message.Text

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println(err)
		return
	}

	switch user.StatusID {
	case db.EnabledStatus:
		{
			SendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case db.WaitingWbState:
		{
			m.getWbStickers(ctx, bot, chatId, message)
		}
	case db.WaitingYaState:
		{
			m.getYandexFbs(ctx, bot, chatId, message)
		}
	case db.WaitingAPI:
		{
			m.changeApi(ctx, bot, chatId, update.Message)
		}
	case db.WaitingSheet:
		{
			m.changeSheet(ctx, bot, chatId, update.Message)
		}
	default:
		log.Println("Такого статуса пользователя нет")
	}

	user.StatusID = db.EnabledStatus
	err = m.repo.UpdateUser(user)
	if err != nil {
		log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен EnabledStatus", chatId)

}

// createStartAdminMarkup создает клавиатуру с кнопками для авторизованного пользователя
func createStartAdminMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Выбери маркетплейс для работы"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: CallbackWbHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: CallbackYandexHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", CallbackData: CallbackOzonHandler})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

// createStartUserMarkup создает клавиатуру с кнопками для неавторизованного пользователя
func createStartUserMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Для доступа к функционалу бота пиши @molchanovz. А пока можешь перейти в наши магазины"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", URL: "https://www.wildberries.ru/seller/27566"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", URL: "https://market.yandex.ru/business--metr-v-kube/3697903"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", URL: "https://www.ozon.ru/seller/metr-v-kube-259267"})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatId int64
	var text string

	if update.Message != nil {
		chatId = update.Message.From.ID
	} else {
		chatId = update.CallbackQuery.From.ID
	}

	_, err := bot.SetMyCommands(ctx, &botlib.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: MessageStartHandler, Description: "Перезапуск бота"},
			{Command: MessageSettingsHandler, Description: "Настройки"},
		},
	})
	if err != nil {
		log.Println("Ошибка регистрации команд:", err)
		return
	}

	_, err = bot.SetChatMenuButton(ctx, &botlib.SetChatMenuButtonParams{
		ChatID: chatId,
		MenuButton: models.MenuButtonCommands{
			Type: models.MenuButtonTypeCommands,
		},
	})
	if err != nil {
		log.Println("Ошибка создания меню: ", err)
		return
	}

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	// если юзера нет - заполняем бд
	if user.TgID == 0 {
		user = &db.User{TgID: chatId, StatusID: db.EnabledStatus, CabinetIDs: make([]int, 0)}
		err := m.repo.CreateUser(user)
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		} else {
			log.Printf("Пользователь %v создан", chatId)
		}
	} else {
		user.StatusID = db.EnabledStatus
		err := m.repo.UpdateUser(user)
		if err != nil {
			log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
		}
		log.Printf("У пользователя %v обновлен EnabledStatus", chatId)
	}

	var startMessage string
	var markup models.InlineKeyboardMarkup

	if user.IsAdmin {
		startMessage, markup = createStartAdminMarkup()
	} else {
		startMessage, markup = createStartUserMarkup()
	}

	if update.Message != nil {
		name := update.Message.From.FirstName
		text = fmt.Sprintf("Привет, %v. %v", name, startMessage)
		_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Println(fmt.Sprintf("ошибка отправки сообщения %v", err))
			return
		}

	} else {
		messageId := update.CallbackQuery.Message.Message.ID
		text = startMessage
		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: messageId, ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Printf("%v", err)
			return
		}
	}
}

func WaitReadyFile(ctx context.Context, bot *botlib.Bot, chatId int64, progressChan chan fbsPrinter.Progress, done chan []string, errChan chan error) error {
	var progressMsgId int
	var lastReportedCurrent int
	var lastTotal int
	var err error
	for {
		select {
		case progress := <-progressChan:
			progressMsgId, err = sendProgress(ctx, bot, chatId, progress, lastReportedCurrent, lastTotal, progressMsgId)
			if err != nil {
				log.Println(err)
				return err
			}

		case filePath := <-done:
			err = sendFiles(ctx, bot, chatId, filePath, progressMsgId)
			if err != nil {
				log.Println(err)
				return err
			}
			return nil

		case err = <-errChan:
			_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: err.Error()})
			if err != nil {
				log.Println(fmt.Sprintf("ошибка отправки сообщения %v", err))
				return err
			}
			return err
		}
	}
}

func sendFiles(ctx context.Context, bot *botlib.Bot, chatId int64, filePath []string, progressMsgId int) error {
	_, err := bot.SendChatAction(ctx, &botlib.SendChatActionParams{
		ChatID: chatId,
		Action: models.ChatActionUploadDocument,
	})
	if err != nil {
		return err
	}

	if len(filePath) == 0 {
		return errors.New("новых заказов нет")
	}

	for _, batchPath := range filePath {
		err = SendMediaMessage(ctx, bot, chatId, batchPath)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if progressMsgId != 0 {
		_, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
			ChatID:    chatId,
			MessageID: progressMsgId,
		})
		if err != nil {
			return err
		}
	}

	text, markup := createStartAdminMarkup()
	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatId,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}
	return nil
}

func sendProgress(ctx context.Context, bot *botlib.Bot, chatId int64, progress fbsPrinter.Progress, lastReportedCurrent int, lastTotal int, progressMsgId int) (int, error) {
	if progress.Current != lastReportedCurrent || progress.Total != lastTotal {
		lastReportedCurrent = progress.Current
		lastTotal = progress.Total

		text := fmt.Sprintf("Обработано заказов: %d из %d", progress.Current, progress.Total)

		if progressMsgId == 0 {
			msg, err := bot.SendMessage(ctx, &botlib.SendMessageParams{
				ChatID: chatId,
				Text:   text,
			})
			if err != nil {
				log.Println(err)
				return 0, err
			} else {
				progressMsgId = msg.ID
			}
		} else {
			_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
				ChatID:    chatId,
				MessageID: progressMsgId,
				Text:      text,
			})
			if err != nil {
				log.Println(err)
				return 0, err
			}
		}
	}
	return progressMsgId, nil
}

type CallbacksForCabinetMarkup struct {
	PaginationCallback string
	SelectCallback     string
	BackCallback       string
}

func createCabinetsMarkup(cabinets []db.Cabinet, callbacks CallbacksForCabinetMarkup, page int, hasNext bool) models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton
	var row []models.InlineKeyboardButton
	var button models.InlineKeyboardButton
	for _, cabinet := range cabinets {
		row = []models.InlineKeyboardButton{}
		button = models.InlineKeyboardButton{Text: cabinet.Name, CallbackData: fmt.Sprintf("%v%v", callbacks.SelectCallback, cabinet.ID)}
		row = append(row, button)

		keyboard = append(keyboard, row)
	}

	//Добавление кнопок для пагинации
	row = []models.InlineKeyboardButton{}
	if page > 1 {
		button = models.InlineKeyboardButton{Text: "⬅️", CallbackData: callbacks.PaginationCallback + fmt.Sprintf("%v", page-1)}
		row = append(row, button)
	}

	if hasNext {
		button = models.InlineKeyboardButton{Text: "➡️", CallbackData: callbacks.PaginationCallback + fmt.Sprintf("%v", page+1)}
		row = append(row, button)
	}

	if row != nil {
		keyboard = append(keyboard, row)
	}

	//row = []models.InlineKeyboardButton{}
	//button = models.InlineKeyboardButton{Text: "Добавить аккаунт", CallbackData: addParserCallback}
	//row = append(row, button)
	//keyboard = append(keyboard, row)

	row = []models.InlineKeyboardButton{}
	button = models.InlineKeyboardButton{Text: "Назад", CallbackData: callbacks.BackCallback}
	row = append(row, button)
	keyboard = append(keyboard, row)

	markup := models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
	return markup
}

func SendTextMessage(ctx context.Context, bot *botlib.Bot, chatId int64, text string) (*models.Message, error) {
	message, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text})
	if err != nil {
		return nil, err
	}
	return message, nil
}

func SendMediaMessage(ctx context.Context, bot *botlib.Bot, chatId int64, filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	inputFile := models.InputFileUpload{
		Filename: filePath,
		Data:     file,
	}

	_, err = bot.SendDocument(ctx, &botlib.SendDocumentParams{ChatID: chatId, Document: &inputFile})
	if err != nil {
		return err
	}
	return nil
}

func initEnv(path, name string) (string, error) {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Ошибка загрузки файла %s: %v\n", path, err)
		return "", fmt.Errorf("ошибка загрузки файла " + path)
	}
	// Получаем значения переменных среды
	env := os.Getenv(name)

	if env == "" {
		return "", fmt.Errorf("переменная среды " + name + " не установлена")
	}
	return env, err
}

// Функция для автоподбора ширины колонок
func autoFitColumns(f *excelize.File, sheet string, columns []string) error {
	for _, col := range columns {
		maxWidth := 8.0 // Минимальная ширина по умолчанию
		rows, err := f.GetRows(sheet)
		if err != nil {
			return err
		}

		// Находим максимальную длину содержимого в колонке
		for _, row := range rows {
			colIdx := int(col[0] - 'A')
			if colIdx < len(row) {
				cellValue := row[colIdx]
				// Учитываем длину текста + 2 символа для отступов
				width := float64(len(cellValue))*1.1 + 2
				if width > maxWidth {
					maxWidth = width
				}
			}
		}

		// Устанавливаем ширину
		if err := f.SetColWidth(sheet, col, col, maxWidth); err != nil {
			return err
		}
	}
	return nil
}

func calculateSmartDemandForecast(salesData []float64) float64 {
	if len(salesData) == 0 {
		return 0
	}

	// Настройки
	const (
		shortWindow = 4  // Анализ последних 4 дней для "горячего" тренда
		longWindow  = 14 // Анализ за 14 дней для базового уровня
	)

	// "горячий" тренд (последние 4 дня)
	hotTrend := 0.0
	if len(salesData) >= shortWindow {
		recent := salesData[len(salesData)-shortWindow:]
		first, last := recent[0], recent[len(recent)-1]
		if first > 0 {
			hotTrend = last / first // Рост в последние дни
		}
	}

	//Среднее за весь период (14 дней)
	fullPeriodAverage := mean(salesData)

	//Среднее за последние 4 дня
	recentAverage := 0.0
	if len(salesData) >= shortWindow {
		recentAverage = mean(salesData[len(salesData)-shortWindow:])
	} else {
		recentAverage = mean(salesData)
	}

	// Динамический вес для тренда
	trendWeight := 0.5  // Базовый вес тренда
	if hotTrend > 2.0 { // Если рост более 2x
		trendWeight = 0.8 // Сильнее учитываем тренд
	}

	// Комбинированный прогноз
	forecast := (recentAverage*trendWeight + fullPeriodAverage*(1-trendWeight)) * float64(longWindow)

	// Гарантия, что прогноз не ниже последних продаж
	if len(salesData) > 0 {
		lastDaySales := salesData[len(salesData)-1]
		minForecast := lastDaySales * float64(longWindow) * 0.7 // Не менее 70% от последнего дня
		if forecast < minForecast {
			forecast = minForecast
		}
	}

	return math.Round(forecast)
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
