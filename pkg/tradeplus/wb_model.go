package tradeplus

import (
	"strings"
	"text/template"
	"time"
	"tradebot/pkg/client/wb"
	"tradebot/pkg/db"
)

type Review struct {
	db.Review
	ArticleDescription *string
}

func NewReview(in *db.Review) *Review {
	if in == nil {
		return nil
	}

	return &Review{
		Review: *in,
	}
}

func (r Review) IsEmpty() bool {
	return (r.Cons == "") && (r.Pros == "") && (r.Text == "")
}

func (r Review) Stars() string {
	if r.Valuation <= 0 {
		return "0"
	}
	stars := strings.Repeat("★", r.Valuation)
	emptyStars := strings.Repeat("☆", 5-r.Valuation)
	return stars + emptyStars
}

func (r Review) ToMessage() string {
	reviewTemplate := `Отзыв на <b>{{.Article}}</b> на {{.Stars}}.` + "\n" +
		`{{if .CustomerName}}<b>Покупатель</b>: {{.CustomerName}}` + "\n" +
		`{{end}}{{if .Pros}}<b>Достоинства</b>: {{.Pros}}` + "\n" +
		`{{end}}{{if .Cons}}<b>Недостатки</b>: {{.Cons}}` + "\n" +
		`{{end}}{{if .Text}}<b>Отзыв</b>: {{.Text}}` + "\n" +
		`{{end}}{{if .Answer}}<b>Ответ</b>: <pre>{{.Answer}}</pre>{{end}}`

	tmpl := template.Must(template.New("review").Parse(reviewTemplate))

	var sb strings.Builder
	err := tmpl.Execute(&sb, r)
	if err != nil {
		return "Ошибка формирования отзыва"
	}

	result := sb.String()
	return strings.TrimSpace(result)
}

func (r Review) ToPrompt() string {
	// add description for LLM
	r.ArticleDescription = SetArticleDescription(r.Article)

	promptTemplate := `Отзыв на {{.Article}} на {{.Valuation}} звезд.
	{{if .ArticleDescription}}Описание товара: {{.ArticleDescription}}
	{{end}}{{if .CustomerName}}Покупатель: {{.CustomerName}}
	{{end}}{{if .Pros}}Достоинства: {{.Pros}}
	{{end}}{{if .Cons}}Недостатки: {{.Cons}}
	{{end}}{{if .Text}}Отзыв: {{.Text}}
	{{end}}{{if .Answer}}Ответ: {{.Answer}}{{end}}`

	tmpl := template.Must(template.New("review").Parse(promptTemplate))

	var sb strings.Builder
	err := tmpl.Execute(&sb, r)
	if err != nil {
		return "Ошибка формирования отзыва"
	}

	result := sb.String()
	return strings.TrimSpace(result)
}

func (r Review) ToDB() *db.Review {
	return &db.Review{
		ID:           r.ID,
		CabinetID:    r.CabinetID,
		ExternalID:   r.ExternalID,
		Text:         r.Text,
		Pros:         r.Pros,
		Cons:         r.Cons,
		Valuation:    r.Valuation,
		Answer:       r.Answer,
		Article:      r.Article,
		CreatedAt:    r.CreatedAt,
		StatusID:     r.StatusID,
		CustomerName: r.CustomerName,
	}
}

func NewReviewFromWB(in wb.Feedback) Review {
	r := db.Review{
		ExternalID:   in.Id,
		Article:      in.ProductDetails.SupplierArticle,
		CustomerName: in.UserName,
		Text:         in.Text,
		Pros:         in.Pros,
		Cons:         in.Cons,
		Valuation:    in.ProductValuation,
	}

	if in.Bables != nil {
		r.Text += "\nПокупатель отметил: " + strings.Join(in.Bables, ", ")
	}

	return Review{
		Review: r,
	}
}

var products = map[string]string{
	"Complekt-Rezinok":                           "Комплект резинок для песочного фильтра",
	"FBW-VM-ALG-1":                               "Альгицид для бассейна против водорослей, грибка и зелени 1 л",
	"FBW-VM-ALG-3":                               "Альгицид для бассейна против водорослей, грибка и зелени 1 л",
	"FBW-VM-ALG-5":                               "Альгицид для бассейна против водорослей, грибка и зелени 1 л",
	"FBW-VM-ECO-ALG-05":                          "Средство для воды от водорослей, грибка и зелени 0,5 л",
	"FBW-VM-ECO-ALG-1":                           "Средство для воды от водорослей, грибка и зелени 1 л",
	"FBW-VM-ECO-ALG-3":                           "Средство для воды от водорослей, грибка и зелени 3 л",
	"FBW-VM-ECO-ALG-5":                           "Средство для воды от водорослей, грибка и зелени 5 л",
	"FBW-VM-ECO-COMPLEX-1kg-20g":                 "Средство 3-в-1 для бассейна в таблетках по 20 г, 1 кг",
	"homut-vm-1":                                 "Хомут для крышки песочного фильтра vm-1 (Деталь №24)",
	"M3AL5":                                      "Альгицид 5 кг для бассейна против водорослей и зелени",
	"Motor_VMFRISBEE":                            "Модуль- мотор (99) управления на робот пылесос",
	"Perehodnik-32-38-mm":                        "Переходник на шланг для песочного фильтра Vommy (Деталь №18)",
	"Perehodnik-i-shlang-38-1m":                  "Переходник для шланга 32/38 мм c шлангом 32 мм 2 м",
	"Schlang32-01":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 1 м",
	"Schlang32-02":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 2 м",
	"Schlang32-03":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 3 м",
	"Schlang32-04":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 4 м",
	"Schlang32-05":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 5 м",
	"Schlang32-06":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 6 м",
	"Schlang32-07":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 7 м",
	"Schlang32-08":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 8 м",
	"Schlang32-09":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 9 м",
	"Schlang32-10":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 10 м",
	"Schlang32-11":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 11 м",
	"Schlang32-12":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 12 м",
	"Schlang32-13":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 13 м",
	"Schlang32-14":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 14 м",
	"Schlang32-15":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 15 м",
	"Schlang32-50":                               "Гофрированный отрезной секционный шланг бассейна 32 мм 50 м",
	"Schlang38-01":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 1 м",
	"Schlang38-02":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 2 м",
	"Schlang38-03":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 3 м",
	"Schlang38-04":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 4 м",
	"Schlang38-05":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 5 м",
	"Schlang38-06":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 6 м",
	"Schlang38-07":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 7 м",
	"Schlang38-08":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 8 м",
	"Schlang38-09":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 9 м",
	"Schlang38-10":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 10 м",
	"Schlang38-11":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 11 м",
	"Schlang38-12":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 12 м",
	"Schlang38-13":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 13 м",
	"Schlang38-14":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 14 м",
	"Schlang38-15":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 15 м",
	"Schlang38-50":                               "Гофрированный отрезной секционный шланг бассейна 38 мм 50 м",
	"SM-DH-10L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 10 л",
	"SM-DH-12L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 12 л",
	"SM-DH-18L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 18 л",
	"SM-DH-30L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 30 л",
	"SM-DH-35L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 35 л",
	"SM-DH-40L":                                  "Осушитель воздуха бытовой для дома, квартиры и подвала 40 л",
	"SM-DH-motor-ventilyatora":                   "Электродвигатель вентилятора для осушителя воздуха Step Mark",
	"SM-DH-Zaglushka_drenazhnogo_otverstiya":     "Заглушка дренажного отверстия для осушителя воздуха Vommy",
	"SMP-DH-10L":                                 "Осушитель воздуха бытовой для дома, квартиры и подвала 10 л",
	"SMP-DH-12L":                                 "Осушитель воздуха бытовой для дома, квартиры и подвала 12 л",
	"VM_AROUND_MAX":                              "Аккумуляторный робот пылесос для бассейна с фильтром 50 м2",
	"VM_HYSON_100P":                              "Ручной вакуумный пылесос для бассейна швабра",
	"VM_HYSON_100P_ruchka":                       "Ручка для ручного пылесоса",
	"VM_HYSON_100P_setka":                        "Сетка для ручного пылесоса Vommy",
	"VM_ROBOT_S2":                                "Аккумуляторный робот пылесос для бассейна с фильтром",
	"VM-1":                                       "Песочный фильтр насос для бассейнов до 18 м3",
	"VM-1 EASY":                                  "Песочный фильтр насос для бассейнов до 14 м3",
	"VM-1 EASY_SPH":                              "Песочный фильтр насос с наполнителем для бассейнов до 14 м3",
	"VM-1_filter_gruboy_ochistki":                "Фильтр грубой очистки для VM-1 и VM-1 Easy (деталь №15)",
	"VM-1_grub_ochistka_23":                      "Крышка фильтра грубой очистки Vommy ДО 2023г (деталь №17)",
	"VM-1_grub_ochistka_24":                      "Крышка фильтра грубой очистки Vommy ОБР. 2024г (деталь №17)",
	"VM-1_kran":                                  "Носик сливного отверстия фильтра Vommy",
	"VM-1_manometr":                              "Манометр для песочного фильтра",
	"vm-1_nabor-rezinok":                         "Набор уплотнительных резинок для песочного фильтра VM-1",
	"VM-1_nasos":                                 "Насос в сборе для песочного фильтра vm-1 (деталь №14)",
	"vm-1_rezinka_bol":                           "Уплотнительная резинка клапанной крышки фильтра (деталь №23)",
	"vm-1_rezinka_mal":                           "Уплотнение крышки фильтра грубой очистки (Деталь №16)",
	"VM-1_rezinka_separatora":                    "Резинка сепаратора под клапанной крышкой для фильтра vm-1",
	"vm-1_Schlang":                               "Шланг для VM-1",
	"VM-1_SPH":                                   "Песочный фильтр насос с наполнителем для бассейнов до 18 м3",
	"VM-1_uplotnitel_32mm_krasniy":               "Уплотнитель №20 для фильтра VM-1. 32 мм красный",
	"VM-1_uplotnitel_38mm_cherniy":               "Уплотнитель №19 для фильтра VM-1. 38 мм черный",
	"VM-3":                                       "Песочный фильтр насос для бассейна до 16 м3, мощность 4 м3/ч",
	"VM-3_krishka":                               "Крышка фильтра VM-3",
	"VM-3_SPH":                                   "Песочный фильтр насос с наполнителем для бассейна до 16 м3",
	"VM-4":                                       "Песочный фильтр насос для бассейна до 16 м3, мощность 4 м3/ч",
	"VM-4_homut":                                 "Хомут для песочного фильтра VM-4",
	"VM-4_krishka_filtra_gruboy_ochistki":        "Крышка фильтра грубой очистки для песочного фильтра vommy",
	"VM-4_manometr":                              "Манометр для песочного фильтра VM-4",
	"VM-4_setka_filtra_gruboy_ochistki":          "Сетка фильтра грубой очистки для VM-4 (деталь №11)",
	"VM-4_SPH":                                   "Песочный фильтр насос с наполнителем для бассейна до 16 м3",
	"VM-4_uplotnitel_filtra":                     "Уплотнитель для фильтра Vommy VM-4 (деталь №12)",
	"vm-4_uplotnitel_klapannoy_kryshki":          "Уплотнительная резинка клапанной крышки фильтра (деталь №7)",
	"VM-5":                                       "Песочный фильтр насос для бассейнов",
	"VM-5_manometr":                              "Манометр для песочного фильтра VM-5",
	"vm-5_nasos":                                 "Насос песочного фильтра для VM-5 (деталь №28)",
	"VM-5_SPH":                                   "Песочный фильтр насос с шариками для бассейнов",
	"VM-5t":                                      "Песочный фильтр насос с таймером для бассейнов",
	"VM-5t_SPH":                                  "Песочный фильтр насос с таймером и шариками для бассейнов",
	"VM-8":                                       "Песочный фильтр-насос для бассейнов",
	"VM-8_bochka":                                "Бочка для песочного фильтра VM-8 (деталь №12)",
	"VM-8_nasos":                                 "Насос для песочного фильтра VM-8",
	"VM-8_slivnaya_krishka_nasosa":               "Сливная крышка для песочного фильтра-насоса для бассейнов",
	"VM-8t":                                      "Песочный фильтр-насос с таймером для бассейнов",
	"VM-ALG-05":                                  "Альгицид против водорослей и зелени 0.5 л",
	"VM-ALG-05x2":                                "Альгицид против водорослей и зелени 1 л",
	"VM-ALG-1":                                   "Альгицид против водорослей и зелени 1 л",
	"VM-ALG-3":                                   "Альгицид против водорослей и зелени 3 л",
	"VM-ALG-5":                                   "Альгицид против водорослей и зелени 5 л",
	"VM-blok_pitaniya":                           "Зарядное устройство для пылесоса Vommy",
	"VM-BLUE-CL-COMPLEX-1kg-20g":                 "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-BLUE-CL-COMPLEX-3kg-20g":                 "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-BLUE-CL-COMPLEX-600g-20g":                "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-BISTRIY-1kg-20gNEW":                   "Хлорные таблетки для бассейна.Быстрый хлор.Химия в бассейн",
	"VM-CL-BISTRIY-3kg-20g":                      "Хлорные таблетки для бассейна.Быстрый хлор.Химия в бассейн",
	"VM-CL-BISTRIY-5kg-20g":                      "Хлорные таблетки для бассейна.Быстрый хлор.Химия в бассейн",
	"VM-CL-COMPLEX-1kg-200g":                     "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-COMPLEX-1kg-20g":                      "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-COMPLEX-3kg-200g":                     "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-COMPLEX-3kg-20g":                      "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-COMPLEX-5kg-200g":                     "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-COMPLEX-5kg-20g":                      "Химия для бассейна в таблетках 3 в 1 хлор,коагулянт,альгицид",
	"VM-CL-MEDLENNIY-1kg-200g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-CL-MEDLENNIY-1kg-20g":                    "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-CL-MEDLENNIY-3kg-200g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-CL-MEDLENNIY-3kg-20g":                    "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-CL-MEDLENNIY-5kg-200g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-CL-MEDLENNIY-5kg-20g":                    "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-Complex-bez-hlora-05-L":                  "Средство химия для бассейна уход 4 в 1 без хлора 0,5 л",
	"VM-Complex-bez-hlora-05х4":                  "Средство химия для бассейна уход 4 в 1 без хлора 0,5 л, 4 шт",
	"VM-Complex-bez-hlora-1-L":                   "Средство химия для бассейна уход 4 в 1 без хлора 1 л",
	"VM-Complex-bez-hlora-3litra":                "Средство химия для бассейна жидкость без хлора уход 4 в 1",
	"VM-Complex-bez-hlora-5litrov":               "Средство химия для бассейна жидкость без хлора уход 4 в 1",
	"VM-dlya_napitkov_1L":                        "Средство от накипи 1 л, химия для чистки кофемашин, чайников",
	"VM-dlya_napitkov_500mL":                     "Средство от накипи 0.5л, химия для чистки кофемашин чайников",
	"VM-Dozator-Bolshoy-Bez-Termometra":          "Поплавок дозатор для бассейна для хлора и химии в таблетках",
	"VM-Dozator-Bolshoy-S-Termometrom":           "Поплавок дозатор для бассейна для химии в таблетках",
	"VM-Dozator-Maliy-Bez-Termometra":            "Поплавок дозатор для бассейна для хлора и химии в таблетках",
	"VM-Dozator-Maliy-S-Termometrom":             "Поплавок дозатор для бассейна для хлора и химии в таблетках",
	"VM-ECO-ALG-05":                              "Альгицид эконом против водорослей и зелени 0,5 л",
	"VM-ECO-ALG-05x2":                            "Альгицид эконом против водорослей и зелени 1 л",
	"VM-ECO-ALG-1":                               "Альгицид эконом против водорослей и зелени 1 л",
	"VM-ECO-ALG-3":                               "Альгицид эконом против водорослей и зелени 3 л",
	"VM-ECO-ALG-5":                               "Альгицид эконом против водорослей и зелени 5 л",
	"VM-ECO-COMPLEX-1kg-200g":                    "Средство 3-в-1 для бассейна в таблетках по 200 г, 1 кг",
	"VM-ECO-COMPLEX-1kg-20g":                     "Средство для бассейна 3 в 1 в таблетках по 20 г, 1 кг",
	"VM-ECO-COMPLEX-3kg-200g":                    "Средство 3-в-1 для бассейна в таблетках по 200 г, 3 кг",
	"VM-ECO-COMPLEX-3kg-20g":                     "Средство для бассейна 3 в 1 в таблетках по 20 г, 3 кг",
	"VM-ECO-COMPLEX-5kg-200g":                    "Средство 3-в-1 для бассейна в таблетках по 200 г, 5 кг",
	"VM-ECO-COMPLEX-5kg-20g":                     "Средство для бассейна 3 в 1 в таблетках по 20 г, 5 кг",
	"VM-ECO-COMPLEX-600g-20g":                    "Средство для бассейна 3 в 1 в таблетках по 20 г, 0,6 кг",
	"VM-ECO-Koagulyant-v-kartrijah":              "Эконом Коагулянт для бассейна картридж 40 таблеток",
	"VM-ECO-MEDLENNIY-1kg-200g":                  "Средство для бассейна , медленный в таблетках по 200 г, 1 кг",
	"VM-ECO-MEDLENNIY-1kg-20g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-ECO-MEDLENNIY-3kg-200g":                  "Средство для бассейна медленный в таблетках по 200 г, 3 кг",
	"VM-ECO-MEDLENNIY-3kg-20g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-ECO-MEDLENNIY-5kg-200g":                  "Средство для бассейна медленный в таблетках по 200 г, 5 кг",
	"VM-ECO-MEDLENNIY-5kg-20g":                   "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-ECO-MEDLENNIY-600g-20g":                  "Химия для бассейна таблетки хлор средство для ухода, очистки",
	"VM-Koagulyant-Granuly-1kg":                  "Коагулянт 1 Химия для бассейна Средство гранулы от мутности",
	"VM-Koagulyant-Granuly-3kg":                  "Коагулянт 3 Химия для бассейна Средство гранулы от мутности",
	"VM-Koagulyant-Granuly-5kg":                  "Коагулянт 5 Химия для бассейна Средство гранулы от мутности",
	"VM-Koagulyant-v-kartrijah":                  "Коагулянт для бассейна картридж 40 таблеток",
	"VM-Koagulyant-v-kartrijah-500g":             "Коагулянт для бассейна картридж 20 таблеток",
	"VM-Koagulyant-zhidkiy-1L":                   "Коагулянт жидкий для бассейна 1 л",
	"VM-Koagulyant-zhidkiy-3L":                   "Коагулянт жидкий для бассейна 3 л",
	"VM-Koagulyant-zhidkiy-5L":                   "Коагулянт жидкий для бассейна 5 л",
	"VM-Meshok":                                  "Мешок для наполнителя песочного фильтра",
	"VM-ot_nakipi_1L":                            "Средство от накипи 1 л, химия для чистки кофемашин, чайников",
	"VM-ot_nakipi_500mL":                         "Средство от накипи 0.5л, химия для чистки кофемашин чайников",
	"VM-pH-Minus-1kg":                            "Средство pH-минус 1 кг порошок для бассейна",
	"VM-pH-Minus-3kg":                            "Средство pH-минус 3 кг порошок для бассейна",
	"VM-pH-Minus-5kg":                            "Средство pH-минус 5 кг порошок для бассейна",
	"VM-Pool-vac":                                "Пылесос донный для бассейна с фильтром",
	"VM-Pool-vac-filter":                         "Сменный многоразовый фильтр для донного пылесоса",
	"VM-Sachok-Classic":                          "Сачок телескопический для чистки бассейна",
	"VM-Sachok-ClassicXL":                        "Сачок телескопический для чистки бассейна",
	"VM-Sachok-Deluxe":                           "Сачок со скребком для чистки бассейна",
	"VM-Sachok-DeluxeXL":                         "Сачок глубокий со скребком для чистки бассейна",
	"VM-Sachok-Glubokiy":                         "Сачок глубокий для очистки бассейна",
	"VM-Sachok-Razborniy-150cm":                  "Сачок неглубокий для очистки бассейна",
	"VM-Shariki":                                 "Наполнитель для песочного фильтра",
	"VM-Shariki-500g":                            "Наполнитель для песочного фильтра 500 грамм",
	"VM-shetka-vacuum-cleaner":                   "Щётка для vacuum-cleaner",
	"VM-ShetkaClassic":                           "Щетка для чистки бассейна донная для пылесоса",
	"VM-ShetkaClassic-Roller":                    "Щетка для чистки и уборки бассейна донная для пылесоса",
	"VM-ShetkaDeluxe":                            "Щетка для чистки бассейна донная для пылесоса",
	"VM-ShetkaOrange":                            "Щетка для чистки бассейна донная для пылесоса",
	"VM-ShetkaPremium":                           "Щетка для чистки бассейна донная для пылесоса",
	"VM-ShetkaPremium-Gibkaya":                   "Щетка для чистки бассейна донная для пылесоса",
	"VM-Shtanga-270cm":                           "Телескопическая штанга для щетки и сачка 270 см",
	"VM-Slivnoy_kran":                            "Сливной кран В СБОРЕ для фильтра Vommy (vm-5/vm-8)",
	"VM-Solar-Shower-20l":                        "Солнечный уличный летний душ для дачи 20 л",
	"VM-Solar-Shower-35l":                        "Солнечный уличный летний душ для дачи 35 л",
	"VM-Solar-Shower-35l_udlinnitelnaya_truba":   "Удлинительная труба для душа 35 л (деталь E)",
	"VM-Solar-Shower-35l_verhn":                  "Верхняя часть для летнего душа 35 л",
	"VM-Solar-Shower-40l":                        "Солнечный уличный летний душ для дачи 40 л",
	"VM-Solar-Shower-60l":                        "Солнечный уличный летний душ для дачи 60 л",
	"VM-Solar-Shower-60l_kran_dlya_nog":          "Нижний кран для душа 60 л",
	"VM-Solar-Shower-60l_leika":                  "Лейка для душа 60 л",
	"VM-Solar-Shower-60l_uplotnitel":             "Уплотнитель для душа 60 л",
	"VM-Termometr-Jimmy-Boy":                     "Термометр плавающий для бассейна",
	"VM-Termometr-Jimmy-Boy-Blue":                "Термометр плавающий для бассейна",
	"VM-termometr-orange":                        "Термометр с зондом для бассейна",
	"VM-Termometr-S-Zondom":                      "Термометр с зондом для бассейна",
	"VM-tester-ph":                               "Тестер для воды бассейна для измерения pH и свободного хлора",
	"VM-vacuum-cleaner":                          "Донный вакуумный ручной пылесос",
	"VM-vacuum-cleaner_meshok":                   "Сменный многоразовый фильтр для донного пылесоса",
	"VM-WINTER-ALG-3":                            "Зимний Консервант против водорослей и зелени 3 л",
	"VM-WINTER-ALG-5":                            "Зимний Консервант против водорослей и зелени 5 л",
	"vm1_detal3":                                 "Крышка слива для песочного фильтр-насоса vommy (деталь №3)",
	"VMBUGSONH3172":                              "Аккумуляторный робот пылесос для бассейна с фильтром",
	"VMBUGSONH3172_charger":                      "Блок питания для робота пылесоса с ЕВРО ВИЛКОЙ",
	"VMFRISBEE_DNISCHE":                          "Нижняя часть (только корпус) робота-пылесоса Vommy",
	"VMFRISBEE1102":                              "Аккумуляторный робот пылесос для бассейна с фильтром",
	"VMFRISBEE1102_akkumulyator":                 "Аккумуляторный блок для робота-пылесоса vommy Around",
	"VMFRISBEE1102_filter":                       "Фильтр для робота Vommy",
	"VMFRISBEE1102_koleso":                       "Колесо для робота Vommy (1 шт)",
	"VMFRISBEE1102_kruk":                         "Ручка-крюк для робота Vommy",
	"VMFRISBEE1102_nizhnyaya_krishka":            "Нижняя часть робота Vommy (без навесных частей)",
	"VMFRISBEE1102_otvetnaya_chast_filter_setki": "Ответная часть фильтр-сетки для робота Vommy",
	"VMFRISBEE1102_pruzhina":                     "Комплект пружин для робота Vommy",
	"VMFRISBEE1102_shetki":                       "Запасные щетки для робота Vommy",
	"VMFRISBEE1102_shtorka":                      "Шторка для робота Vommy",
	"VMFRISBEE1102_shtorki":                      "Шторки для робота Vommy",
	"VMFRISBEE1102_zacshyolka":                   "Защёлка для робота-пылесоса Vommy Around",
	"VMFRISBEE1102_ZAGLUSKA":                     "Заглушка для зарядки робот - пылесос Vommy",
	"VMMIA1005":                                  "Проводной робот пылесос для бассейна с фильтром для ухода",
	"VMOPTIMUS2052":                              "Беспроводной робот-пылесос для бассейна с резиновым валиком",
	"VMOPTIMUSCOMP":                              "Беспроводной робот-пылесос для бассейна с пенным валиком",
	"VMOPTIMUSPRO":                               "Робот-пылесос для бассейна с дистанционным управлением",
	"VMTYPHOR1":                                  "Беспроводной робот пылесос для бассейна с фильтром для ухода",
	"VMW-026":                                    "Усиленная лестница для каркасного бассейна intex bestwa",
}

func SetArticleDescription(article string) *string {
	if v, ok := products[article]; ok {
		return &v
	}
	return nil
}

func NewReviewsFromWB(in *wb.Review) Reviews {
	if in == nil {
		return nil
	}

	var reviews = make(Reviews, 0, len(in.Data.Feedbacks))
	for i := range in.Data.Feedbacks {
		review := NewReviewFromWB(in.Data.Feedbacks[i])
		review.StatusID = db.ReviewStatusCompleted
		reviews = append(reviews, review)
	}
	return reviews
}

type ReviewWB struct {
	wb.Review
}

type Card struct {
	NmID        int
	ImtID       int
	NmUUID      string
	SubjectID   int
	SubjectName string
	VendorCode  string
	Brand       string
	Title       string
	Description string
	NeedKiz     bool
	Dimensions  struct {
		Width        int
		Height       int
		Length       int
		WeightBrutto float64
		IsValid      bool
	}
	Characteristics []struct {
		Id    int
		Name  string
		Value interface{}
	}
	Sizes []struct {
		ChrtID   int
		TechSize string
		WbSize   string
		Skus     []string
	}
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCardList(in *wb.CardList) Cards {
	if in == nil {
		return nil
	}

	cards := make([]Card, 0, len(in.Cards))
	for i := range in.Cards {
		c := Card{
			NmID:            in.Cards[i].NmID,
			ImtID:           in.Cards[i].ImtID,
			NmUUID:          in.Cards[i].NmUUID,
			SubjectID:       in.Cards[i].SubjectID,
			SubjectName:     in.Cards[i].SubjectName,
			VendorCode:      in.Cards[i].VendorCode,
			Brand:           in.Cards[i].Brand,
			Title:           in.Cards[i].Title,
			Description:     in.Cards[i].Description,
			NeedKiz:         in.Cards[i].NeedKiz,
			Characteristics: nil,
			Sizes:           nil,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
		c.Dimensions.Height = in.Cards[i].Dimensions.Height
		c.Dimensions.Length = in.Cards[i].Dimensions.Length
		c.Dimensions.Width = in.Cards[i].Dimensions.Width
		c.Dimensions.IsValid = in.Cards[i].Dimensions.IsValid
		c.Dimensions.WeightBrutto = in.Cards[i].Dimensions.WeightBrutto

		cards = append(cards, c)
	}

	return cards
}

type Return struct {
	Barcode          string `json:"barcode"`
	Brand            string `json:"brand"`
	CompletedDt      string `json:"completedDt"`
	DstOfficeAddress string `json:"dstOfficeAddress"`
	DstOfficeId      int    `json:"dstOfficeId"`
	ExpiredDt        string `json:"expiredDt"`
	IsStatusActive   int    `json:"isStatusActive"`
	NmId             int    `json:"nmId"`
	OrderDt          string `json:"orderDt"`
	OrderId          int    `json:"orderId"`
	ReadyToReturnDt  string `json:"readyToReturnDt"`
	Reason           string `json:"reason"`
	ReturnType       string `json:"returnType"`
	ShkId            int64  `json:"shkId"`
	Srid             string `json:"srid"`
	Status           string `json:"status"`
	StickerId        string `json:"stickerId"`
	SubjectName      string `json:"subjectName"`
	TechSize         string `json:"techSize"`
}

func NewReturns(in *wb.ReturnList) []Return {
	if in == nil {
		return nil
	}

	returns := make([]Return, 0, len(in.Report))
	for i := range in.Report {
		returns = append(returns, Return{
			Barcode:          in.Report[i].Barcode,
			Brand:            in.Report[i].Brand,
			CompletedDt:      in.Report[i].CompletedDt,
			DstOfficeAddress: in.Report[i].DstOfficeAddress,
			DstOfficeId:      in.Report[i].DstOfficeId,
			ExpiredDt:        in.Report[i].ExpiredDt,
			IsStatusActive:   in.Report[i].IsStatusActive,
			NmId:             in.Report[i].NmId,
			OrderDt:          in.Report[i].OrderDt,
			OrderId:          in.Report[i].OrderId,
			ReadyToReturnDt:  in.Report[i].ReadyToReturnDt,
			Reason:           in.Report[i].Reason,
			ReturnType:       in.Report[i].ReturnType,
			ShkId:            in.Report[i].ShkId,
			Srid:             in.Report[i].Srid,
			Status:           in.Report[i].Status,
			StickerId:        in.Report[i].StickerId,
			SubjectName:      in.Report[i].SubjectName,
			TechSize:         in.Report[i].TechSize,
		})
	}

	return returns
}
