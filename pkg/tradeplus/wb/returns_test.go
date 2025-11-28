package wb

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/go-pg/pg/v10"
	"github.com/vmkteam/vfs"
	"testing"
	"tradebot/pkg/db"
)

var (
	testRepo db.TradebotRepo
	cabinet  *db.Cabinet
	err      error
	cfg      Config
	ctx      = context.Background()
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

func TestMain(m *testing.M) {
	if _, err = toml.DecodeFile("/Users/sergey/GolandProjects/tradebot/cfg/local.toml", &cfg); err != nil {
		return
	}

	pgdb := pg.Connect(cfg.Database)
	dbc := db.New(pgdb)
	testRepo = db.NewTradebotRepo(dbc)

	cabinet, err = testRepo.CabinetByID(ctx, 9)
	if err != nil {
		return
	}
	m.Run()
}

func TestReturnsManager_WriteReturns(t *testing.T) {
	m := NewReturnsManager(cabinet.Key)
	_, err = m.WriteReturns()
}
