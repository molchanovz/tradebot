package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"strings"
	"tradebot/pkg/db"
	"tradebot/pkg/fbsPrinter"
	"tradebot/pkg/marketplaces/YANDEX"
)

const (
	MarketYandex                       = "YANDEX"
	CallbackYandexHandler              = MarketYandex
	CallbackYandexStickersHandler      = "YANDEX-STICKERS_"
	CallbackYandexOrdersHandler        = "YANDEX-ORDERS_"
	CallbackYandexCabinetsHandler      = "YANDEX-CABINETS"
	CallbackSelectYandexCabinetHandler = "CABINET-YANDEX_"
)

func (m *Manager) yandexHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	cabinets, err := m.repo.GetCabinets(MarketYandex)
	if err != nil {
		log.Println(err)
		return
	}

	text := "Выберите кабинет"
	callbacks := CallbacksForCabinetMarkup{
		PaginationCallback: CallbackYandexCabinetsHandler,
		SelectCallback:     CallbackSelectYandexCabinetHandler,
		BackCallback:       CallbackStartHandler,
	}
	markup := createCabinetsMarkup(cabinets, callbacks, 0, false)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}
func (m *Manager) yandexCabinetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]

	text := "Кабинет Яндекс"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: fmt.Sprintf("%v+%v", CallbackYandexStickersHandler, cabinetId)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: fmt.Sprintf("%v+%v", CallbackYandexOrdersHandler, cabinetId)})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	user.StatusID = db.WaitingYaState
	err = m.repo.UpdateUser(user)
	if err != nil {
		log.Println("Ошибка обновления WaitingYaState пользователя: ", err)
	}

	text := fmt.Sprintf("Отправь мне номер отгрузки")
	var buttonBack []models.InlineKeyboardButton

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: update.CallbackQuery.Message.Message.ID, ChatID: chatId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) yandexOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	cabinets, err := m.repo.GetCabinets(MarketYandex)
	if err != nil {
		log.Println(err)
		return
	}

	err = YANDEX.NewService(cabinets...).GetOrdersAndReturnsManager().WriteToGoogleSheets()
	if err != nil {
		log.Printf("%v", err)
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return
	}

	_, err = SendTextMessage(ctx, bot, chatId, "Заказы яндекс за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	done := make(chan []string)
	progressChan := make(chan fbsPrinter.Progress)
	errChan := make(chan error)
	var filePaths []string

	cabinets, err := m.repo.GetCabinets(MarketYandex)
	if err != nil {
		log.Println(err)
		return
	}

	var cabinetFBS db.Cabinet

	for _, c := range cabinets {
		if c.Type == "fbs" {
			cabinetFBS = c
			break
		}
	}

	go func() {
		filePath, err := YANDEX.NewService(cabinetFBS).GetStickersFbsManager().GetOrdersInfo(supplyId, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			errChan <- err
			return
		}

		filePaths = append(filePaths, filePath)

		done <- filePaths
	}()

	err = WaitReadyFile(ctx, bot, chatId, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			return
		}
		return
	}
}
