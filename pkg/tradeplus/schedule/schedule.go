package schedule

import (
	"context"
	"fmt"
	"github.com/vmkteam/embedlog"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"
	"tradebot/pkg/tradeplus/yandex"
)

type Manager struct {
	embedlog.Logger
	tm *tradeplus.Manager
}

func NewManager(dbc db.DB, logger embedlog.Logger) Manager {
	return Manager{tm: tradeplus.NewManager(dbc), Logger: logger}
}

func (s *Manager) WriteWB(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		return fmt.Errorf("fetch cabinet failed: %w", err)
	}

	err = wb.NewService(cabinets[0]).GetOrdersManager().Write()
	if err != nil {
		return fmt.Errorf("write orders failed: %w", err)
	}

	return nil
}

func (s *Manager) WriteOzon(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketOzon)
	if err != nil {
		return err
	}

	titleRange := "!A1"
	fbsRange := "!A2:B1000"
	fboRange := "!D2:E1000"
	returnsRange := "!G2:H1000"

	maxValuesCount, err := ozon.NewService(cabinets[0]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		return err
	}

	maxValuesCount += 3
	titleRange = fmt.Sprintf("!A%v", maxValuesCount)

	maxValuesCount++
	fbsRange = fmt.Sprintf("!A%v:B%v", maxValuesCount, maxValuesCount+1000)
	fboRange = fmt.Sprintf("!D%v:E%v", maxValuesCount, maxValuesCount+1000)
	returnsRange = fmt.Sprintf("!G%v:H%v", maxValuesCount, maxValuesCount+1000)

	_, err = ozon.NewService(cabinets[1]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		return err
	}

	return nil
}

func (s *Manager) WriteYandex(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketYandex)
	if err != nil {
		return err
	}

	err = yandex.NewService(cabinets...).GetOrdersAndReturnsManager().Write()
	if err != nil {
		return err
	}

	return nil
}

func (s *Manager) ClearOrders(ctx context.Context) error {
	return s.tm.DeleteOrders(ctx)
}
