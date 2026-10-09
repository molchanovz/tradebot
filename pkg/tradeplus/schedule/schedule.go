package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tradebot/pkg/bot"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"
	"tradebot/pkg/tradeplus/yandex"

	"github.com/vmkteam/embedlog"
)

type Manager struct {
	embedlog.Logger
	tm *tradeplus.Manager
	bs *bot.Service
}

func NewManager(dbc db.DB, logger embedlog.Logger, bs *bot.Service) Manager {
	return Manager{tm: tradeplus.NewManager(dbc), Logger: logger, bs: bs}
}

func (s *Manager) WriteWB(ctx context.Context) error {
	return s.writeOrders(ctx, db.MarketWB)
}

func (s *Manager) WriteOzon(ctx context.Context) error {
	return s.writeOrders(ctx, db.MarketOzon)
}

func (s *Manager) WriteOzonShipments(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketOzon)
	if err != nil {
		return err
	}

	msk := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(msk).AddDate(0, 0, -tradeplus.OrdersDaysAgo)
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, msk)

	var failed []string
	for _, cab := range cabinets {
		if cab.Settings.ShipmentsSheetID == "" {
			s.Print(ctx, fmt.Sprintf("ozonShipments: cabinet=%d skipped (no shipmentsSheetId)", cab.ID))
			continue
		}
		m, err := ozon.NewShipmentsManager(cab)
		if err != nil {
			failed = append(failed, fmt.Sprintf("cab=%d init: %v", cab.ID, err))
			continue
		}
		if err := m.WriteForDate(ctx, yesterday); err != nil {
			failed = append(failed, err.Error())
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("ozonShipments: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (s *Manager) WriteOzonShipmentsAll(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketOzon)
	if err != nil {
		return err
	}

	msk := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(msk).AddDate(0, 0, -tradeplus.OrdersDaysAgo)
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, msk)

	var failed []string
	for _, cab := range cabinets {
		if cab.Settings.ShipmentsAllSheetID == "" {
			continue
		}
		m, err := ozon.NewShipmentsManager(cab)
		if err != nil {
			failed = append(failed, fmt.Sprintf("cab=%d init: %v", cab.ID, err))
			continue
		}
		if err := m.WriteAggregatedForDate(ctx, yesterday); err != nil {
			failed = append(failed, err.Error())
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("ozonShipmentsAll: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (s *Manager) WriteWBShipments(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		return err
	}

	msk := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(msk).AddDate(0, 0, -tradeplus.OrdersDaysAgo)
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, msk)

	var failed []string
	for _, cab := range cabinets {
		if cab.Settings.ShipmentsSheetID == "" {
			s.Print(ctx, fmt.Sprintf("wbShipments: cabinet=%d skipped (no shipmentsSheetId)", cab.ID))
			continue
		}
		m, err := wb.NewShipmentsManager(cab)
		if err != nil {
			failed = append(failed, fmt.Sprintf("cab=%d init: %v", cab.ID, err))
			continue
		}
		if err := m.WriteForDate(ctx, yesterday); err != nil {
			failed = append(failed, err.Error())
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("wbShipments: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (s *Manager) WriteWBShipmentsAll(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		return err
	}

	msk := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(msk).AddDate(0, 0, -tradeplus.OrdersDaysAgo)
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, msk)

	var failed []string
	for _, cab := range cabinets {
		if cab.Settings.ShipmentsAllSheetID == "" {
			continue
		}
		m, err := wb.NewShipmentsManager(cab)
		if err != nil {
			failed = append(failed, fmt.Sprintf("cab=%d init: %v", cab.ID, err))
			continue
		}
		if err := m.WriteAggregatedForDate(ctx, yesterday); err != nil {
			failed = append(failed, err.Error())
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("wbShipmentsAll: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (s *Manager) WriteYandexShipments(ctx context.Context) error {
	cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketYandex)
	if err != nil {
		return err
	}

	msk := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(msk).AddDate(0, 0, -tradeplus.OrdersDaysAgo)
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, msk)

	var failed []string
	for _, cab := range cabinets {
		if cab.Type != "fbs" {
			continue
		}
		if cab.Settings.ShipmentsSheetID == "" {
			s.Print(ctx, fmt.Sprintf("ymShipments: cabinet=%d skipped (no shipmentsSheetId)", cab.ID))
			continue
		}
		m, err := yandex.NewShipmentsManager(ctx, cab)
		if err != nil {
			failed = append(failed, fmt.Sprintf("cab=%d init: %v", cab.ID, err))
			continue
		}
		if err := m.WriteForDate(ctx, yesterday); err != nil {
			failed = append(failed, err.Error())
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("ymShipments: %s", strings.Join(failed, "; "))
	}
	return nil
}

func (s *Manager) WriteYandex(ctx context.Context) error {
	return s.writeOrders(ctx, db.MarketYandex)
}

// writeOrders заносит в таблицу заказов маркетплейса дни этого месяца по вчера,
// которых в ней ещё нет (см. tradeplus.OrdersDays).
func (s *Manager) writeOrders(ctx context.Context, mp string) error {
	written, err := s.bs.Manager().SyncOrders(ctx, mp)
	if len(written) > 0 {
		s.Print(ctx, fmt.Sprintf("%s orders written: %s", mp, tradeplus.FormatDays(written)))
	}
	if err != nil {
		return fmt.Errorf("write %s orders: %w", mp, err)
	}
	return nil
}

func (s *Manager) ClearOrders(ctx context.Context) error {
	return s.tm.DeleteOrders(ctx)
}

func (s *Manager) FetchReviews(ctx context.Context) error {
	return s.bs.Manager().FetchReviews(ctx)
}

func (s *Manager) ProcessReviews(ctx context.Context) error {
	return s.bs.Manager().ProcessReviews(ctx)
}
