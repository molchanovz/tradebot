package wb

import (
	"github.com/BurntSushi/toml"
	"github.com/openai/openai-go/v3"
	"github.com/stretchr/testify/require"
	"testing"
	"tradebot/pkg/client/openAI"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/test"
)

var (
	cfg  test.Config
	oaim = openAI.NewManager(cfg.OpenAI.Token)
)

func TestMain(m *testing.M) {
	if _, err := toml.DecodeFile("/Users/sergey/GolandProjects/tradebot/cfg/local.toml", &cfg); err != nil {
		return
	}
	m.Run()
}

func TestReviewManager_Reviews(t *testing.T) {
	dbc, err := test.Setup()
	repo := db.NewTradebotRepo(dbc)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cabinet, err := repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: openai.Ptr("WB")})
	require.NoError(t, err)

	m := NewReviewManager(*dbc, tradeplus.NewCabinet(cabinet), oaim)
	reviews, err := m.Reviews(t.Context())
	require.NoError(t, err)
	t.Log(reviews)
}

//func TestReviewManager_AnswerReview(t *testing.T) {
//	err := test.Setup()
//	require.NoError(t, err)
//	require.NotNil(t, test.Cfg)
//
//	cabinet, err := test.Repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: tradeplus.Ptr("WB")})
//	require.NoError(t, err)
//
//	m := NewReviewManager(cabinet.Key)
//
//	err = m.AnswerReview()
//	require.NoError(t, err)
//}
