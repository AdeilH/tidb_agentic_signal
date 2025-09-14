package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/trader"
)

type MarketDataService struct {
	binanceClient   *trader.Client
	wsHub           *trader.WSHub
	wsManager       *trader.BinanceWebSocketManager
	marketDataStore *db.MarketDataStore
	symbols         []string
	mu              sync.RWMutex
	priceCache      map[string]PriceData
	running         bool
	cancel          context.CancelFunc
}

type PriceData struct {
	Symbol             string    `json:"symbol"`
	Price              float64   `json:"price"`
	PriceChange        float64   `json:"priceChange"`
	PriceChangePercent float64   `json:"priceChangePercent"`
	Volume             float64   `json:"volume"`
	QuoteVolume        float64   `json:"quoteVolume"`
	High               float64   `json:"high"`
	Low                float64   `json:"low"`
	OpenPrice          float64   `json:"openPrice"`
	BidPrice           float64   `json:"bidPrice"`
	AskPrice           float64   `json:"askPrice"`
	Timestamp          time.Time `json:"timestamp"`
}

type MarketUpdate struct {
	Type      string                 `json:"type"`
	Symbol    string                 `json:"symbol"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type OrderBookData struct {
	Symbol    string     `json:"symbol"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
	Timestamp time.Time  `json:"timestamp"`
}

type TradeData struct {
	Symbol       string    `json:"symbol"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	TradeTime    time.Time `json:"tradeTime"`
	IsBuyerMaker bool      `json:"isBuyerMaker"`
}

type KlineData struct {
	Symbol    string    `json:"symbol"`
	Interval  string    `json:"interval"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	OpenTime  time.Time `json:"openTime"`
	CloseTime time.Time `json:"closeTime"`
	IsClosed  bool      `json:"isClosed"`
}

func NewMarketDataService(binanceClient *trader.Client, wsHub *trader.WSHub, database *db.DB) *MarketDataService {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}

	service := &MarketDataService{
		binanceClient:   binanceClient,
		wsHub:           wsHub,
		wsManager:       trader.NewBinanceWebSocketManager(binanceClient, symbols),
		marketDataStore: db.NewMarketDataStore(database),
		symbols:         symbols,
		priceCache:      make(map[string]PriceData),
	}

	// Set up WebSocket data handlers
	service.setupWebSocketHandlers()

	return service
}

// setupWebSocketHandlers configures the WebSocket data handlers
func (s *MarketDataService) setupWebSocketHandlers() {
	// Ticker data handler
	s.wsManager.SetDataHandler("ticker", func(data interface{}) {
		if tickerEvent, ok := data.(trader.WSTickerEvent); ok {
			s.handleTickerUpdate(tickerEvent)
		}
	})

	// Trade data handler
	s.wsManager.SetDataHandler("trade", func(data interface{}) {
		if tradeEvent, ok := data.(trader.WSTradeEvent); ok {
			s.handleTradeUpdate(tradeEvent)
		}
	})

	// Depth data handler
	s.wsManager.SetDataHandler("depth", func(data interface{}) {
		if depthEvent, ok := data.(trader.WSDepthEvent); ok {
			s.handleDepthUpdate(depthEvent)
		}
	})

	// Kline data handler
	s.wsManager.SetDataHandler("kline", func(data interface{}) {
		if klineEvent, ok := data.(trader.WSKlineEvent); ok {
			s.handleKlineUpdate(klineEvent)
		}
	})
}

func (s *MarketDataService) StartStreaming(ctx context.Context) error {
	if s.running {
		return fmt.Errorf("market data service is already running")
	}

	var streamCtx context.Context
	streamCtx, s.cancel = context.WithCancel(ctx)
	s.running = true

	log.Println("🚀 Starting enhanced market data streaming service...")

	// Start the new WebSocket manager with production endpoints
	if err := s.wsManager.Start(streamCtx); err != nil {
		s.running = false
		return fmt.Errorf("failed to start WebSocket manager: %w", err)
	}

	// Start background services
	go s.updatePriceCache(streamCtx)
	go s.persistMarketData(streamCtx)

	log.Printf("✅ Market data service started successfully for %d symbols", len(s.symbols))
	return nil
}

func (s *MarketDataService) Stop() {
	if !s.running {
		return
	}

	log.Println("Stopping Market Data Service...")

	// Stop WebSocket manager
	if s.wsManager != nil {
		s.wsManager.Stop()
	}

	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
	log.Println("Market Data Service stopped")
}

func (s *MarketDataService) IsRunning() bool {
	return s.running
}

func (s *MarketDataService) GetPriceData(symbol string) (PriceData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.priceCache[symbol]
	return data, exists
}

func (s *MarketDataService) GetAllPrices() map[string]PriceData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]PriceData)
	for k, v := range s.priceCache {
		result[k] = v
	}
	return result
}

func (s *MarketDataService) AddSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.symbols {
		if existing == symbol {
			return // Already exists
		}
	}

	s.symbols = append(s.symbols, symbol)
	log.Printf("Added symbol %s to market data service", symbol)
}

func (s *MarketDataService) updatePriceCache(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fetchAndCacheAllPrices()
		}
	}
}

func (s *MarketDataService) fetchAndCacheAllPrices() {
	tickers, err := s.binanceClient.GetTicker24hr("")
	if err != nil {
		log.Printf("Error fetching 24hr tickers: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ticker := range tickers {
		// Only cache symbols we're tracking
		isTracked := false
		for _, symbol := range s.symbols {
			if ticker.Symbol == symbol {
				isTracked = true
				break
			}
		}

		if !isTracked {
			continue
		}

		price, _ := strconv.ParseFloat(ticker.LastPrice, 64)
		priceChange, _ := strconv.ParseFloat(ticker.PriceChange, 64)
		priceChangePercent, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)
		volume, _ := strconv.ParseFloat(ticker.Volume, 64)
		quoteVolume, _ := strconv.ParseFloat(ticker.QuoteVolume, 64)
		high, _ := strconv.ParseFloat(ticker.HighPrice, 64)
		low, _ := strconv.ParseFloat(ticker.LowPrice, 64)
		openPrice, _ := strconv.ParseFloat(ticker.OpenPrice, 64)
		bidPrice, _ := strconv.ParseFloat(ticker.BidPrice, 64)
		askPrice, _ := strconv.ParseFloat(ticker.AskPrice, 64)

		s.priceCache[ticker.Symbol] = PriceData{
			Symbol:             ticker.Symbol,
			Price:              price,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			Volume:             volume,
			QuoteVolume:        quoteVolume,
			High:               high,
			Low:                low,
			OpenPrice:          openPrice,
			BidPrice:           bidPrice,
			AskPrice:           askPrice,
			Timestamp:          time.Now(),
		}
	}

	log.Printf("Updated price cache for %d symbols", len(s.priceCache))
}

func (s *MarketDataService) startTickerStream(ctx context.Context, symbol string) error {
	return s.binanceClient.SubscribeTickerStream(ctx, symbol, func(event trader.WSTickerEvent) {
		price, _ := strconv.ParseFloat(event.LastPrice, 64)
		priceChange, _ := strconv.ParseFloat(event.PriceChange, 64)
		priceChangePercent, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
		volume, _ := strconv.ParseFloat(event.TotalTradedVolume, 64)
		quoteVolume, _ := strconv.ParseFloat(event.TotalTradedQuote, 64)
		high, _ := strconv.ParseFloat(event.HighPrice, 64)
		low, _ := strconv.ParseFloat(event.LowPrice, 64)
		openPrice, _ := strconv.ParseFloat(event.OpenPrice, 64)
		bidPrice, _ := strconv.ParseFloat(event.BestBidPrice, 64)
		askPrice, _ := strconv.ParseFloat(event.BestAskPrice, 64)

		priceData := PriceData{
			Symbol:             event.Symbol,
			Price:              price,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			Volume:             volume,
			QuoteVolume:        quoteVolume,
			High:               high,
			Low:                low,
			OpenPrice:          openPrice,
			BidPrice:           bidPrice,
			AskPrice:           askPrice,
			Timestamp:          time.UnixMilli(event.EventTime),
		}

		// Update cache
		s.mu.Lock()
		s.priceCache[event.Symbol] = priceData
		s.mu.Unlock()

		// Broadcast to frontend
		update := MarketUpdate{
			Type:      "price_update",
			Symbol:    event.Symbol,
			Data:      priceData,
			Timestamp: time.Now(),
		}

		s.wsHub.Broadcast("market_update", update)
	})
}

func (s *MarketDataService) startTradeStream(ctx context.Context, symbol string) error {
	return s.binanceClient.SubscribeTradeStream(ctx, symbol, func(event trader.WSTradeEvent) {
		price, _ := strconv.ParseFloat(event.Price, 64)
		quantity, _ := strconv.ParseFloat(event.Quantity, 64)

		tradeData := TradeData{
			Symbol:       event.Symbol,
			Price:        price,
			Quantity:     quantity,
			TradeTime:    time.UnixMilli(event.TradeTime),
			IsBuyerMaker: event.IsBuyerMaker,
		}

		update := MarketUpdate{
			Type:      "trade",
			Symbol:    event.Symbol,
			Data:      tradeData,
			Timestamp: time.Now(),
		}

		s.wsHub.Broadcast("market_update", update)
	})
}

func (s *MarketDataService) startDepthStream(ctx context.Context, symbol string) error {
	return s.binanceClient.SubscribeDepthStream(ctx, symbol, 20, func(event trader.WSDepthEvent) {
		orderBookData := OrderBookData{
			Symbol:    event.Symbol,
			Bids:      event.Bids,
			Asks:      event.Asks,
			Timestamp: time.UnixMilli(event.EventTime),
		}

		update := MarketUpdate{
			Type:      "depth",
			Symbol:    event.Symbol,
			Data:      orderBookData,
			Timestamp: time.Now(),
		}

		s.wsHub.Broadcast("market_update", update)
	})
}

func (s *MarketDataService) startKlineStream(ctx context.Context, symbol, interval string) error {
	return s.binanceClient.SubscribeKlineStream(ctx, symbol, interval, func(event trader.WSKlineEvent) {
		open, _ := strconv.ParseFloat(event.Kline.Open, 64)
		high, _ := strconv.ParseFloat(event.Kline.High, 64)
		low, _ := strconv.ParseFloat(event.Kline.Low, 64)
		close, _ := strconv.ParseFloat(event.Kline.Close, 64)
		volume, _ := strconv.ParseFloat(event.Kline.Volume, 64)

		klineData := KlineData{
			Symbol:    event.Symbol,
			Interval:  event.Kline.Interval,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			OpenTime:  time.UnixMilli(event.Kline.StartTime),
			CloseTime: time.UnixMilli(event.Kline.EndTime),
			IsClosed:  event.Kline.IsClosed,
		}

		update := MarketUpdate{
			Type:      "kline",
			Symbol:    event.Symbol,
			Data:      klineData,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"interval": interval,
				"closed":   event.Kline.IsClosed,
			},
		}

		s.wsHub.Broadcast("market_update", update)
	})
}

// GetMarketSummary returns a summary of current market conditions
func (s *MarketDataService) GetMarketSummary() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalVolume, avgPriceChange float64
	gainers := make([]PriceData, 0)
	losers := make([]PriceData, 0)

	for _, data := range s.priceCache {
		totalVolume += data.QuoteVolume
		avgPriceChange += data.PriceChangePercent

		if data.PriceChangePercent > 5.0 {
			gainers = append(gainers, data)
		} else if data.PriceChangePercent < -5.0 {
			losers = append(losers, data)
		}
	}

	if len(s.priceCache) > 0 {
		avgPriceChange /= float64(len(s.priceCache))
	}

	return map[string]interface{}{
		"total_symbols":    len(s.priceCache),
		"total_volume":     totalVolume,
		"avg_price_change": avgPriceChange,
		"major_gainers":    gainers,
		"major_losers":     losers,
		"last_updated":     time.Now(),
		"service_running":  s.running,
	}
}

// persistMarketData handles storing real-time market data to TiDB
func (s *MarketDataService) persistMarketData(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Store data every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeCachedPricesToTiDB()
		}
	}
}

// storeCachedPricesToTiDB stores current price cache to TiDB
func (s *MarketDataService) storeCachedPricesToTiDB() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for symbol, priceData := range s.priceCache {
		marketPrice := db.MarketPrice{
			Symbol:             symbol,
			Price:              priceData.Price,
			PriceChange:        &priceData.PriceChange,
			PriceChangePercent: &priceData.PriceChangePercent,
			Volume:             &priceData.Volume,
			QuoteVolume:        &priceData.QuoteVolume,
			High24h:            &priceData.High,
			Low24h:             &priceData.Low,
			OpenPrice:          &priceData.OpenPrice,
			BidPrice:           &priceData.BidPrice,
			AskPrice:           &priceData.AskPrice,
			Timestamp:          priceData.Timestamp,
		}

		if err := s.marketDataStore.StorePrice(marketPrice); err != nil {
			log.Printf("Error storing price for %s: %v", symbol, err)
		}
	}
}

// StoreTradeData stores individual trade data to TiDB
func (s *MarketDataService) StoreTradeData(symbol string, price, quantity float64, tradeTime time.Time, isBuyerMaker bool) error {
	trade := db.MarketTrade{
		Symbol:       symbol,
		Price:        price,
		Quantity:     quantity,
		TradeTime:    tradeTime,
		IsBuyerMaker: isBuyerMaker,
		Timestamp:    time.Now(),
	}
	return s.marketDataStore.StoreTrade(trade)
}

// StoreKlineData stores candlestick data to TiDB
func (s *MarketDataService) StoreKlineData(symbol, interval string, open, high, low, close, volume float64, openTime, closeTime time.Time, isClosed bool) error {
	kline := db.MarketKline{
		Symbol:       symbol,
		IntervalType: interval,
		OpenPrice:    open,
		HighPrice:    high,
		LowPrice:     low,
		ClosePrice:   close,
		Volume:       volume,
		OpenTime:     openTime,
		CloseTime:    closeTime,
		IsClosed:     isClosed,
		Timestamp:    time.Now(),
	}
	return s.marketDataStore.StoreKline(kline)
}

// StoreOrderBookData stores order book snapshot to TiDB
func (s *MarketDataService) StoreOrderBookData(symbol string, bids, asks [][]string) error {
	orderbook := db.MarketOrderBook{
		Symbol:     symbol,
		Bids:       bids,
		Asks:       asks,
		DepthLevel: len(bids), // Use actual depth
		Timestamp:  time.Now(),
	}
	return s.marketDataStore.StoreOrderBook(orderbook)
}

// GetMarketDataFromTiDB retrieves stored market data for analysis
func (s *MarketDataService) GetMarketDataFromTiDB() map[string]interface{} {
	// Get latest prices from TiDB
	prices, err := s.marketDataStore.GetLatestPrices()
	if err != nil {
		log.Printf("Error fetching prices from TiDB: %v", err)
		return nil
	}

	// Convert to response format
	result := make(map[string]interface{})
	priceMap := make(map[string]db.MarketPrice)

	for _, price := range prices {
		priceMap[price.Symbol] = price
	}

	result["prices"] = priceMap
	result["total_symbols"] = len(prices)
	result["data_source"] = "TiDB"
	result["last_updated"] = time.Now()

	return result
}

// GetTradingSignalsFromTiDB analyzes stored data for trading decisions
func (s *MarketDataService) GetTradingSignalsFromTiDB(symbol string) map[string]interface{} {
	signals := make(map[string]interface{})

	// Get price history for trend analysis
	priceHistory, err := s.marketDataStore.GetPriceHistory(symbol, 50)
	if err != nil {
		log.Printf("Error fetching price history: %v", err)
		return signals
	}

	// Get recent trades for volume analysis
	recentTrades, err := s.marketDataStore.GetRecentTrades(symbol, 100)
	if err != nil {
		log.Printf("Error fetching recent trades: %v", err)
		return signals
	}

	// Get volume metrics
	volumeMetrics, err := s.marketDataStore.GetTradingVolume(symbol, 24)
	if err != nil {
		log.Printf("Error fetching volume metrics: %v", err)
		return signals
	}

	// Simple trend analysis
	if len(priceHistory) >= 2 {
		latest := priceHistory[0].Price
		previous := priceHistory[1].Price
		priceDirection := "SIDEWAYS"

		changePercent := ((latest - previous) / previous) * 100
		if changePercent > 2.0 {
			priceDirection = "BULLISH"
		} else if changePercent < -2.0 {
			priceDirection = "BEARISH"
		}

		signals["price_direction"] = priceDirection
		signals["price_change_percent"] = changePercent
	}

	// Volume analysis
	if volumeRatio, exists := volumeMetrics["volume_ratio"]; exists && volumeRatio > 0 {
		if volumeRatio > 0.6 {
			signals["volume_sentiment"] = "BULLISH"
		} else if volumeRatio < 0.4 {
			signals["volume_sentiment"] = "BEARISH"
		} else {
			signals["volume_sentiment"] = "NEUTRAL"
		}
	}

	signals["symbol"] = symbol
	signals["price_data_points"] = len(priceHistory)
	signals["trade_data_points"] = len(recentTrades)
	signals["volume_metrics"] = volumeMetrics
	signals["analysis_time"] = time.Now()

	return signals
}

// WebSocket handler methods for the new WebSocket manager

func (s *MarketDataService) handleTickerUpdate(event trader.WSTickerEvent) {
	price, _ := strconv.ParseFloat(event.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(event.PriceChange, 64)
	priceChangePercent, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
	volume, _ := strconv.ParseFloat(event.TotalTradedVolume, 64)
	quoteVolume, _ := strconv.ParseFloat(event.TotalTradedQuote, 64)
	high, _ := strconv.ParseFloat(event.HighPrice, 64)
	low, _ := strconv.ParseFloat(event.LowPrice, 64)
	openPrice, _ := strconv.ParseFloat(event.OpenPrice, 64)
	bidPrice, _ := strconv.ParseFloat(event.BestBidPrice, 64)
	askPrice, _ := strconv.ParseFloat(event.BestAskPrice, 64)

	priceData := PriceData{
		Symbol:             event.Symbol,
		Price:              price,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		High:               high,
		Low:                low,
		OpenPrice:          openPrice,
		BidPrice:           bidPrice,
		AskPrice:           askPrice,
		Timestamp:          time.Now(),
	}

	s.mu.Lock()
	s.priceCache[event.Symbol] = priceData
	s.mu.Unlock()

	// Store to TiDB
	marketPrice := db.MarketPrice{
		Symbol:             event.Symbol,
		Price:              price,
		PriceChange:        &priceChange,
		PriceChangePercent: &priceChangePercent,
		Volume:             &volume,
		QuoteVolume:        &quoteVolume,
		High24h:            &high,
		Low24h:             &low,
		OpenPrice:          &openPrice,
		BidPrice:           &bidPrice,
		AskPrice:           &askPrice,
		Timestamp:          time.Now(),
	}

	err := s.marketDataStore.StorePrice(marketPrice)
	if err != nil {
		log.Printf("Error storing price data for %s: %v", event.Symbol, err)
	}

	// Broadcast to WebSocket clients
	update := MarketUpdate{
		Type:      "ticker",
		Symbol:    event.Symbol,
		Data:      priceData,
		Timestamp: time.Now(),
	}
	s.wsHub.Broadcast("market_update", update)

	log.Printf("💰 %s: $%s (%.2f%%)", event.Symbol, event.LastPrice, priceChangePercent)
}

func (s *MarketDataService) handleTradeUpdate(event trader.WSTradeEvent) {
	price, _ := strconv.ParseFloat(event.Price, 64)
	quantity, _ := strconv.ParseFloat(event.Quantity, 64)
	tradeTime := time.Unix(0, event.TradeTime*int64(time.Millisecond))

	// Store to TiDB
	err := s.StoreTradeData(event.Symbol, price, quantity, tradeTime, event.IsBuyerMaker)
	if err != nil {
		log.Printf("Error storing trade data for %s: %v", event.Symbol, err)
	}

	// Broadcast to WebSocket clients
	update := MarketUpdate{
		Type:      "trade",
		Symbol:    event.Symbol,
		Data:      event,
		Timestamp: time.Now(),
	}
	s.wsHub.Broadcast("market_update", update)

	log.Printf("⚡ %s Trade: %s @ %s", event.Symbol, event.Quantity, event.Price)
}

func (s *MarketDataService) handleDepthUpdate(event trader.WSDepthEvent) {
	// Store to TiDB
	err := s.StoreOrderBookData(event.Symbol, event.Bids, event.Asks)
	if err != nil {
		log.Printf("Error storing order book data for %s: %v", event.Symbol, err)
	}

	// Broadcast to WebSocket clients
	update := MarketUpdate{
		Type:      "depth",
		Symbol:    event.Symbol,
		Data:      event,
		Timestamp: time.Now(),
	}
	s.wsHub.Broadcast("market_update", update)

	log.Printf("📊 %s Order Book: %d bids, %d asks", event.Symbol, len(event.Bids), len(event.Asks))
}

func (s *MarketDataService) handleKlineUpdate(event trader.WSKlineEvent) {
	if !event.Kline.IsClosed {
		return // Only process closed klines
	}

	open, _ := strconv.ParseFloat(event.Kline.Open, 64)
	high, _ := strconv.ParseFloat(event.Kline.High, 64)
	low, _ := strconv.ParseFloat(event.Kline.Low, 64)
	close, _ := strconv.ParseFloat(event.Kline.Close, 64)
	volume, _ := strconv.ParseFloat(event.Kline.Volume, 64)

	openTime := time.Unix(0, event.Kline.StartTime*int64(time.Millisecond))
	closeTime := time.Unix(0, event.Kline.EndTime*int64(time.Millisecond))

	// Store to TiDB
	err := s.StoreKlineData(event.Kline.Symbol, event.Kline.Interval, open, high, low, close, volume, openTime, closeTime, event.Kline.IsClosed)
	if err != nil {
		log.Printf("Error storing kline data for %s: %v", event.Kline.Symbol, err)
	}

	// Broadcast to WebSocket clients
	update := MarketUpdate{
		Type:      "kline",
		Symbol:    event.Kline.Symbol,
		Data:      event,
		Timestamp: time.Now(),
	}
	s.wsHub.Broadcast("market_update", update)

	log.Printf("📈 %s Kline [%s]: O=%s H=%s L=%s C=%s V=%s",
		event.Kline.Symbol, event.Kline.Interval,
		event.Kline.Open, event.Kline.High, event.Kline.Low, event.Kline.Close, event.Kline.Volume)
}
