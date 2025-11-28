package openAI

import (
	"github.com/BurntSushi/toml"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
	"testing"
)

type Config struct {
	OpenAI struct {
		Token string
	}
}

var (
	configPath = "../../../cfg/local.toml"
	cfg        Config
)

func setup() error {
	_, err := toml.DecodeFile(configPath, &cfg)
	return err
}

func TestNewService(t *testing.T) {
	err := setup()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	service := NewManager(cfg.OpenAI.Token)

	Convey("Channels", t, func() {
		var request = `
Ты — автоответчик компании. Твоя задача — кратко и вежливо отвечать на отзывы покупателей.

Требования:
1. Приветствуй покупателя по имени и всегда благодари за отзыв.
2. Пиши от лица компании (мы/нас).
3. Ответ <= 150 символов. Жёстко соблюдай.
4. Никакой "воды" — только по делу.
5. Ничего не выдумывай, опирайся только на данные из отзыва.
6. Рекомендации давай только для положительных отзывов.
7. Если рекомендация неуместна — не упоминай её.
8. Не упоминай артикулы, если их нет в списке рекомендаций.
9. Запрещено указывать или описывать товар, если покупатель сам его не назвал.

Рекомендации:
- Коагулянт — 123123

Отзыв на VM-Sachok-Razborniy-150cm на 3 звезд.
Покупатель: Елизавета
Отзыв: Товар не соответствует цене качество на очень низком уровне, ручка сильно гнется.
Покупатель отметил:`
		answer, err := service.AnswerReview(t.Context(), request)
		So(err, ShouldBeNil)
		So(answer, ShouldNotBeNil)
		t.Log(answer)
	})
}
