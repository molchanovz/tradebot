package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/yandex"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	CallbackYandexHandler              = "YANDEX"
	CallbackYandexStickersHandler      = "YANDEX-STICKERS_"
	CallbackYandexOrdersHandler        = "YANDEX-ORDERS_"
	CallbackYandexCabinetsHandler      = "YANDEX-CABINETS"
	CallbackSelectYandexCabinetHandler = "CABINET-YANDEX_"
)

func (m *Manager) yandexHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketYandex)
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

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) yandexCabinetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetID := parts[1]

	text := "Кабинет Яндекс"

	var buttonsRow []models.InlineKeyboardButton
	var allButtons [][]models.InlineKeyboardButton

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: fmt.Sprintf("%v%v", CallbackYandexStickersHandler, cabinetID)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackYandexHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	user, err := m.bl.UserByChatID(ctx, chatID)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	_, err = m.bl.SetUserStatus(ctx, user, db.StatusWaitingYaState)
	if err != nil {
		log.Println("Ошибка обновления StatusWaitingYaState пользователя: ", err)
		return
	}

	text := "Отправь мне номер отгрузки"
	var buttonBack []models.InlineKeyboardButton

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: update.CallbackQuery.Message.Message.ID, ChatID: chatID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) yandexOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketYandex)
	if err != nil {
		log.Println(err)
		return
	}

	err = yandex.NewService(cabinets...).GetOrdersAndReturnsManager().Write()
	if err != nil {
		log.Printf("%v", err)
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return
	}

	_, err = SendTextMessage(ctx, bot, chatID, "Заказы яндекс за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) getYandexFbs(ctx context.Context, bot *botlib.Bot, chatID int64, supplyID string) {
	done := make(chan []string)
	progressChan := make(chan tradeplus.Progress)
	errChan := make(chan error)
	var filePaths []string

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketYandex)
	if err != nil {
		log.Println(err)
		return
	}

	var cabinetFBS tradeplus.Cabinet

	for _, c := range cabinets {
		if c.Type == "fbs" {
			cabinetFBS = c
			break
		}
	}

	go func() {
		filePath, err := yandex.NewService(cabinetFBS).GetStickersFbsManager().GetOrdersInfo(supplyID, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			errChan <- err
			return
		}

		filePaths = append(filePaths, filePath)

		done <- filePaths
	}()

	err = WaitReadyFile(ctx, bot, chatID, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
		if err != nil {
			return
		}
		return
	}
}
