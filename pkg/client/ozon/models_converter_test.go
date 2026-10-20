package ozon

import (
	"context"
	"os"
	"testing"

	"tradebot/pkg/db"

	"github.com/BurntSushi/toml"
	"github.com/go-pg/pg/v10"
	"github.com/vmkteam/vfs"
)

type Config struct {
	Database *pg.Options
	Server   struct {
		Host      string
		Port      int
		IsDevel   bool
		EnableVFS bool
	}
	Sentry struct {
		Environment string
		DSN         string
	}
	VFS vfs.Config
}

var (
	testRepo db.TradebotRepo
	cabinet  *db.Cabinet
	cfg      Config
	ctx      = context.Background()
)

// TestMain готовит кабинет 3 из локальной БД для интеграционных тестов, которые ходят
// в API Озона. Без локального конфига (например, в CI) кабинета нет и такие тесты
// не нужны, а остальные тесты пакета работают как обычно.
func TestMain(m *testing.M) {
	if _, err := toml.DecodeFile("/Users/sergey/GolandProjects/tradebot/cfg/local.toml", &cfg); err == nil {
		testRepo = db.NewTradebotRepo(db.New(pg.Connect(cfg.Database)))
		cabinet, _ = testRepo.CabinetByID(ctx, 3)
	}
	os.Exit(m.Run())
}
