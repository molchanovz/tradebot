package wb

import (
	"context"
	"fmt"
	"tradebot/pkg/client/chatgptsrv"
	"tradebot/pkg/client/wb"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
)

const Prompt = `
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

Отзыв на VM-1 на 4 звезд.
Покупатель: Никита
Отзыв: Фильтр неплох, но можно было лучше.
Покупатель отметил:`

type ReviewManager struct {
	dbc     db.DB
	repo    db.TradebotRepo
	client  wb.Client
	chatgpt *chatgptsrv.Client
	cabinet *tradeplus.Cabinet
}

func NewReviewManager(dbc db.DB, cabinet *tradeplus.Cabinet, chatgpt *chatgptsrv.Client) ReviewManager {
	return ReviewManager{
		dbc:     dbc,
		repo:    db.NewTradebotRepo(dbc),
		client:  wb.NewClient(cabinet.Key),
		chatgpt: chatgpt,
		cabinet: cabinet,
	}
}

func (m ReviewManager) Reviews(ctx context.Context) ([]tradeplus.Review, error) {
	reviews, err := m.client.Reviews()
	if err != nil {
		return nil, err
	}

	unansweredReviews := tradeplus.NewReviewsFromWB(reviews)
	externalIDs := unansweredReviews.UniqueExternalIDs()
	existsReviews, err := m.repo.ReviewsByFilters(ctx, &db.ReviewSearch{ExternalIDs: externalIDs}, db.PagerNoLimit)
	if err != nil {
		return nil, err
	}

	var newReviews = make([]tradeplus.Review, 0)

	externalIDx := tradeplus.NewReviews(existsReviews).IndexByExternalID()

	answer, err := m.chatgpt.Chatgpt.Send(context.Background(), Prompt)
	if err != nil {
		return nil, err
	}
	fmt.Println("answer: ", answer)

	for _, nr := range unansweredReviews {
		if _, ok := externalIDx[nr.ExternalID]; ok {
			continue
		}
		//var answer string
		//if !nr.IsEmpty() {
		//	request := Prompt + nr.ToMessage()
		//	answer, err = m.chatgpt.Chatgpt.Send(ctx, request)
		//	if err != nil {
		//		return nil, err
		//	}
		//}
		//
		//nr.Answer = answer
		//nr.CabinetID = m.cabinet.ID
		//
		//_, err = m.repo.AddReview(ctx, nr.ToDB())
		//if err != nil {
		//	return nil, err
		//}
		//
		//newReviews = append(newReviews, nr)
	}

	return newReviews, nil
}

func (m ReviewManager) AnswerReview(ctx context.Context, reviewId string) error {
	review, err := m.repo.OneReview(ctx, &db.ReviewSearch{ExternalID: &reviewId}, db.WithColumns(db.Columns.Review.Answer))
	if err != nil {
		return err
	}

	err = m.client.AnswerReview(reviewId, review.Answer)
	if err != nil {
		return err
	}

	return nil
}
