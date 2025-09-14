package svc

import (
	"sync"

	"github.com/adeilh/agentic_go_signals/internal/api"
	"github.com/adeilh/agentic_go_signals/internal/config"
	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/kimi"
	"github.com/adeilh/agentic_go_signals/internal/trader"
)

var (
	once          sync.Once
	DB            *db.DB
	KClient       *kimi.Client
	BinanceClient *trader.Client
	App           *api.App
)

func Init(cfg *config.Config) error {
	var initErr error
	once.Do(func() {
		var err error
		DB, err = db.Open(cfg.DBDSN)
		if err != nil {
			initErr = err
			return
		}
		if err := db.AutoMigrate(DB); err != nil {
			initErr = err
			return
		}
		KClient = kimi.NewClient(cfg.KimiKey)
		BinanceClient = trader.NewClientWithConfig(cfg.BinanceKey, cfg.BinanceSecret, cfg.BinanceProduction)
		App = api.New(DB, BinanceClient, KClient)
	})
	return initErr
}
