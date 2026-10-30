package ozon

import (
	"context"
	"os"
	"testing"
	"time"

	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus/test"

	"github.com/BurntSushi/toml"
	"github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"
)

var (
	testRepo db.TradebotRepo
	cabinet  *db.Cabinet
	err      error
	cfg      = test.Cfg
	ctx      = context.Background()
)

// TestMain готовит кабинет 3 из локальной БД для интеграционных тестов, которые ходят
// в API Озона. Без локального конфига (например, в CI) кабинета нет и такие тесты
// пропускаются через ozonCabinet, а остальные тесты пакета работают как обычно.
func TestMain(m *testing.M) {
	if _, err := toml.DecodeFile("/Users/sergey/GolandProjects/tradebot/cfg/local.toml", &cfg); err == nil {
		testRepo = db.NewTradebotRepo(db.New(pg.Connect(cfg.Database)))
		cabinet, _ = testRepo.CabinetByID(ctx, 3)
	}
	os.Exit(m.Run())
}

// ozonCabinet возвращает кабинет для интеграционного теста или пропускает тест.
func ozonCabinet(t *testing.T) *db.Cabinet {
	t.Helper()
	if testing.Short() || cabinet == nil || cabinet.ClientID == nil {
		t.Skip("нужны кабинет 3 в локальной БД и доступ к API Озона")
	}
	return cabinet
}

func TestOrdersManager_GetReturnsMap(t *testing.T) {
	cabinet := ozonCabinet(t)
	m := NewOrdersManager(*cabinet.ClientID, cabinet.Key)
	since := time.Now().AddDate(0, 0, -2).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, -1).Format("2006-01-02") + "T21:00:00.000Z"

	got, err := m.GetReturnsMap(*cabinet.ClientID, cabinet.Key, since, to)
	require.NoError(t, err)
	t.Log(got)
}
