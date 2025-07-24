package scheduler

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"log/slog"
	"net/http"
	"os"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"
	"tradebot/pkg/tradeplus/yandex"

	"github.com/vmkteam/cron"
)

type Service struct {
	tm *tradeplus.Manager
	cm *cron.Manager
}

func NewService(dbc db.DB) Service {
	return Service{tm: tradeplus.NewManager(dbc), cm: cron.NewManager()}
}

func (s *Service) Start(ctx context.Context) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stderr)

	sl := NewLogger(false)

	s.cm.Use(
		cron.WithMetrics("test"),
		cron.WithDevel(false),
		cron.WithSLog(sl),
		cron.WithLogger(log.Printf, "test-run"),
		cron.WithMaintenance(log.Printf),
		cron.WithSkipActive(),
		cron.WithRecover(), // recover() inside
		cron.WithSentry(),  // recover() inside
	)

	// add simple func
	var ordersExpression cron.Schedule = "37 11 * * *"
	s.cm.AddFunc("wbOrders", ordersExpression, s.WriteWB())
	s.cm.AddFunc("ozonOrders", ordersExpression, s.WriteOzon())
	s.cm.AddFunc("yandexOrders", ordersExpression, s.WriteYandex())

	// run cron
	if err := s.cm.Run(ctx); err != nil {
		sl.Error(ctx, err.Error())
	}

	// print schedule (two variants)
	s.cm.TextSchedule(os.Stdout)
	sl.Print(ctx, "cron initialized", "job", s.cm.State())
	sl.Print(ctx, "open this url for cron ui", "url", "http://localhost:2112/debug/cron")
	sl.Print(ctx, "open this url for metrics", "url", "http://localhost:2112/metrics")

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/debug/cron", s.cm.Handler)
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		sl.Error(ctx, err.Error())
	}
}

func (s *Service) WriteWB() cron.Func {
	return func(ctx context.Context) error {
		cabinets, err := s.tm.GetCabinetsByMp(ctx, db.MarketWB)
		if err != nil {
			return err
		}

		err = wb.NewService(cabinets[0]).GetOrdersManager().Write()
		if err != nil {
			return err
		}

		return nil
	}
}

func (s *Service) WriteOzon() cron.Func {
	return func(ctx context.Context) error {
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
}

func (s *Service) WriteYandex() cron.Func {
	return func(ctx context.Context) error {
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
}

// Logger is a simple text/json Slog Logger.
type Logger struct {
	*slog.Logger
}

func NewLogger(json bool) Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if json {
		l = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return Logger{
		Logger: l,
	}
}

func (l Logger) Print(ctx context.Context, msg string, args ...any) {
	l.InfoContext(ctx, msg, args...)
}
func (l Logger) Error(ctx context.Context, msg string, args ...any) {
	l.ErrorContext(ctx, msg, args...)
}
