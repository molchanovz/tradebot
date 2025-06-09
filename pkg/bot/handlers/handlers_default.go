package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"log"
	"math"
	"os"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/WB"
	"tradebot/pkg/marketplaces/YANDEX"
)

const (
	CallbackStartHandler = "START"

	CallbackWbHandler                = "WB"
	CallbackYandexHandler            = "YANDEX"
	CallbackOzonHandler              = "OZON"
	CallbackWbFbsHandler             = "WB-FBS"
	CallbackYandexFbsHandler         = "YANDEX-FBS"
	CallbackWbOrdersHandler          = "WB-ORDERS"
	CallbackYandexOrdersHandler      = "YANDEX-ORDERS"
	CallbackOzonOrdersHandler        = "OZON-ORDERS_"
	CallbackOzonStocksHandler        = "OZON-STOCKS_"
	CallbackWbStocksHandler          = "WB-STOCKS"
	CallbackOzonStickersHandler      = "OZON-STICKERS_"
	CallbackOzonPrintStickersHandler = "OZON-PRINT-STICKERS_"
	CallbackClustersHandler          = "OZON-CLUSTERS"
	CallbackOzonCabinetsHandler      = "OZON-CABINETS"
	CallbackSelectCabinetHandler     = "CABINET_"
)

type Manager struct {
	b             *botlib.Bot
	db            *gorm.DB
	ozonService   OZON.Service
	wbService     WB.Service
	yandexService YANDEX.Service
	myChatId      string
}

func NewBotManager(wbService WB.Service, yandexService YANDEX.Service, db *gorm.DB, myChatId string) *Manager {
	return &Manager{
		wbService:     wbService,
		yandexService: yandexService,
		db:            db,
		myChatId:      myChatId,
	}
}

func (m *Manager) SetBot(bot *botlib.Bot) {
	m.b = bot
}
func (m *Manager) GetBot() *botlib.Bot {
	return m.b
}

func (m *Manager) RegisterBotHandlers() {
	m.b.RegisterHandler(botlib.HandlerTypeMessageText, "/start", botlib.MatchTypePrefix, m.startHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackStartHandler, botlib.MatchTypePrefix, m.startHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbHandler, botlib.MatchTypeExact, wbHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexHandler, botlib.MatchTypeExact, m.yandexHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonHandler, botlib.MatchTypeExact, m.ozonHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbFbsHandler, botlib.MatchTypeExact, m.wbFbsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexFbsHandler, botlib.MatchTypeExact, m.yandexFbsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbOrdersHandler, botlib.MatchTypePrefix, m.wbOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexOrdersHandler, botlib.MatchTypePrefix, m.yandexOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonOrdersHandler, botlib.MatchTypePrefix, m.ozonOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStocksHandler, botlib.MatchTypePrefix, m.ozonStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbStocksHandler, botlib.MatchTypePrefix, wbStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStickersHandler, botlib.MatchTypePrefix, m.ozonStickersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonPrintStickersHandler, botlib.MatchTypePrefix, m.ozonPrintStickers)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackClustersHandler, botlib.MatchTypePrefix, m.ozonClustersHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackSelectCabinetHandler, botlib.MatchTypePrefix, m.ozonCabinetHandler)

	//b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypePrefix, wbOrdersHandler)

}

func (m *Manager) DefaultHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	message := update.Message.Text

	var user db.User
	// Смотрим есть ли артикул в бд
	result := m.db.Where(`"tgId" = ?`, chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	switch user.StatusId {
	case db.EnabledStatus:
		{
			sendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case db.WaitingWbState:
		{
			m.getWbFbs(ctx, bot, chatId, message)
		}
	case db.WaitingYaState:
		{
			m.getYandexFbs(ctx, bot, chatId, message)
		}
	default:
		panic("unhandled default case")
	}

	err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
		TgId:     chatId,
		StatusId: db.EnabledStatus,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен EnabledStatus", chatId)

}

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

	var user db.User
	// Смотрим есть ли юзер в бд
	result := m.db.Where(`"tgId" = ?`, chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding chatId: ", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.TgId == 0 {
		user = db.User{TgId: chatId, StatusId: db.EnabledStatus}
		err := m.db.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
			StatusId: db.EnabledStatus,
		}).Error
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
			log.Printf("%v", err)
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

func sendTextMessage(ctx context.Context, bot *botlib.Bot, chatId int64, text string) (*models.Message, error) {
	message, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text})
	if err != nil {
		return nil, err
	}
	return message, nil
}

func sendMediaMessage(ctx context.Context, bot *botlib.Bot, chatId int64, filePath string) error {

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
