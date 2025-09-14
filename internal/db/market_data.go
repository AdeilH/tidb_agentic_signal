package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// MarketDataStore handles market data persistence in TiDB
type MarketDataStore struct {
	db *DB
}

func NewMarketDataStore(db *DB) *MarketDataStore {
	return &MarketDataStore{db: db}
}

// MarketPrice represents the market_prices table structure
type MarketPrice struct {
	ID                 int64     `json:"id"`
	Symbol             string    `json:"symbol"`
	Price              float64   `json:"price"`
	PriceChange        *float64  `json:"price_change"`
	PriceChangePercent *float64  `json:"price_change_percent"`
	Volume             *float64  `json:"volume"`
	QuoteVolume        *float64  `json:"quote_volume"`
	High24h            *float64  `json:"high_24h"`
	Low24h             *float64  `json:"low_24h"`
	OpenPrice          *float64  `json:"open_price"`
	BidPrice           *float64  `json:"bid_price"`
	AskPrice           *float64  `json:"ask_price"`
	Timestamp          time.Time `json:"timestamp"`
}

// MarketOrderBook represents the market_orderbook table structure
type MarketOrderBook struct {
	ID         int64      `json:"id"`
	Symbol     string     `json:"symbol"`
	Bids       [][]string `json:"bids"`
	Asks       [][]string `json:"asks"`
	DepthLevel int        `json:"depth_level"`
	Timestamp  time.Time  `json:"timestamp"`
}

// MarketTrade represents the market_trades table structure
type MarketTrade struct {
	ID           int64     `json:"id"`
	Symbol       string    `json:"symbol"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	TradeTime    time.Time `json:"trade_time"`
	IsBuyerMaker bool      `json:"is_buyer_maker"`
	Timestamp    time.Time `json:"timestamp"`
}

// MarketKline represents the market_klines table structure
type MarketKline struct {
	ID           int64     `json:"id"`
	Symbol       string    `json:"symbol"`
	IntervalType string    `json:"interval_type"`
	OpenPrice    float64   `json:"open_price"`
	HighPrice    float64   `json:"high_price"`
	LowPrice     float64   `json:"low_price"`
	ClosePrice   float64   `json:"close_price"`
	Volume       float64   `json:"volume"`
	QuoteVolume  *float64  `json:"quote_volume"`
	OpenTime     time.Time `json:"open_time"`
	CloseTime    time.Time `json:"close_time"`
	IsClosed     bool      `json:"is_closed"`
	TradeCount   *int      `json:"trade_count"`
	Timestamp    time.Time `json:"timestamp"`
}

// MarketSummary represents the market_summary table structure
type MarketSummary struct {
	ID              int64     `json:"id"`
	Symbol          string    `json:"symbol"`
	AvgPrice        *float64  `json:"avg_price"`
	Volume24h       *float64  `json:"volume_24h"`
	PriceTrend      string    `json:"price_trend"` // BULLISH, BEARISH, SIDEWAYS
	Volatility      *float64  `json:"volatility"`
	SupportLevel    *float64  `json:"support_level"`
	ResistanceLevel *float64  `json:"resistance_level"`
	Timestamp       time.Time `json:"timestamp"`
}

// StorePrice stores real-time price data
func (m *MarketDataStore) StorePrice(price MarketPrice) error {
	query := `INSERT INTO market_prices (
		symbol, price, price_change, price_change_percent, volume, quote_volume,
		high_24h, low_24h, open_price, bid_price, ask_price, ts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.db.conn.Exec(query,
		price.Symbol, price.Price, price.PriceChange, price.PriceChangePercent,
		price.Volume, price.QuoteVolume, price.High24h, price.Low24h,
		price.OpenPrice, price.BidPrice, price.AskPrice, price.Timestamp,
	)
	return err
}

// StoreOrderBook stores order book snapshot
func (m *MarketDataStore) StoreOrderBook(orderbook MarketOrderBook) error {
	bidsJSON, err := json.Marshal(orderbook.Bids)
	if err != nil {
		return fmt.Errorf("failed to marshal bids: %w", err)
	}

	asksJSON, err := json.Marshal(orderbook.Asks)
	if err != nil {
		return fmt.Errorf("failed to marshal asks: %w", err)
	}

	query := `INSERT INTO market_orderbook (symbol, bids, asks, depth_level, ts) VALUES (?, ?, ?, ?, ?)`
	_, err = m.db.conn.Exec(query, orderbook.Symbol, bidsJSON, asksJSON, orderbook.DepthLevel, orderbook.Timestamp)
	return err
}

// StoreTrade stores individual trade data
func (m *MarketDataStore) StoreTrade(trade MarketTrade) error {
	query := `INSERT INTO market_trades (symbol, price, quantity, trade_time, is_buyer_maker, ts) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.db.conn.Exec(query, trade.Symbol, trade.Price, trade.Quantity, trade.TradeTime, trade.IsBuyerMaker, trade.Timestamp)
	return err
}

// StoreKline stores candlestick data with upsert logic
func (m *MarketDataStore) StoreKline(kline MarketKline) error {
	query := `INSERT INTO market_klines (
		symbol, interval_type, open_price, high_price, low_price, close_price,
		volume, quote_volume, open_time, close_time, is_closed, trade_count, ts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		high_price = GREATEST(high_price, VALUES(high_price)),
		low_price = LEAST(low_price, VALUES(low_price)),
		close_price = VALUES(close_price),
		volume = VALUES(volume),
		quote_volume = VALUES(quote_volume),
		close_time = VALUES(close_time),
		is_closed = VALUES(is_closed),
		trade_count = VALUES(trade_count),
		ts = VALUES(ts)`

	_, err := m.db.conn.Exec(query,
		kline.Symbol, kline.IntervalType, kline.OpenPrice, kline.HighPrice,
		kline.LowPrice, kline.ClosePrice, kline.Volume, kline.QuoteVolume,
		kline.OpenTime, kline.CloseTime, kline.IsClosed, kline.TradeCount, kline.Timestamp,
	)
	return err
}

// StoreSummary stores market summary and analysis
func (m *MarketDataStore) StoreSummary(summary MarketSummary) error {
	query := `INSERT INTO market_summary (
		symbol, avg_price, volume_24h, price_trend, volatility,
		support_level, resistance_level, ts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.db.conn.Exec(query,
		summary.Symbol, summary.AvgPrice, summary.Volume24h, summary.PriceTrend,
		summary.Volatility, summary.SupportLevel, summary.ResistanceLevel, summary.Timestamp,
	)
	return err
}

// GetLatestPrices retrieves latest prices for all symbols
func (m *MarketDataStore) GetLatestPrices() ([]MarketPrice, error) {
	query := `SELECT DISTINCT 
		symbol, price, price_change, price_change_percent, volume, quote_volume,
		high_24h, low_24h, open_price, bid_price, ask_price, ts
	FROM market_prices p1
	WHERE ts = (SELECT MAX(ts) FROM market_prices p2 WHERE p2.symbol = p1.symbol)
	ORDER BY symbol`

	rows, err := m.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []MarketPrice
	for rows.Next() {
		var p MarketPrice
		err := rows.Scan(&p.Symbol, &p.Price, &p.PriceChange, &p.PriceChangePercent,
			&p.Volume, &p.QuoteVolume, &p.High24h, &p.Low24h, &p.OpenPrice,
			&p.BidPrice, &p.AskPrice, &p.Timestamp)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

// GetSymbolPrice retrieves latest price for a specific symbol
func (m *MarketDataStore) GetSymbolPrice(symbol string) (*MarketPrice, error) {
	query := `SELECT symbol, price, price_change, price_change_percent, volume, quote_volume,
		high_24h, low_24h, open_price, bid_price, ask_price, ts
	FROM market_prices 
	WHERE symbol = ? 
	ORDER BY ts DESC 
	LIMIT 1`

	var p MarketPrice
	err := m.db.conn.QueryRow(query, symbol).Scan(
		&p.Symbol, &p.Price, &p.PriceChange, &p.PriceChangePercent,
		&p.Volume, &p.QuoteVolume, &p.High24h, &p.Low24h, &p.OpenPrice,
		&p.BidPrice, &p.AskPrice, &p.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPriceHistory retrieves price history for technical analysis
func (m *MarketDataStore) GetPriceHistory(symbol string, limit int) ([]MarketPrice, error) {
	query := `SELECT symbol, price, price_change, price_change_percent, volume, quote_volume,
		high_24h, low_24h, open_price, bid_price, ask_price, ts
	FROM market_prices 
	WHERE symbol = ? 
	ORDER BY ts DESC 
	LIMIT ?`

	rows, err := m.db.conn.Query(query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []MarketPrice
	for rows.Next() {
		var p MarketPrice
		err := rows.Scan(&p.Symbol, &p.Price, &p.PriceChange, &p.PriceChangePercent,
			&p.Volume, &p.QuoteVolume, &p.High24h, &p.Low24h, &p.OpenPrice,
			&p.BidPrice, &p.AskPrice, &p.Timestamp)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

// GetRecentTrades retrieves recent trades for volume analysis
func (m *MarketDataStore) GetRecentTrades(symbol string, limit int) ([]MarketTrade, error) {
	query := `SELECT symbol, price, quantity, trade_time, is_buyer_maker, ts
	FROM market_trades 
	WHERE symbol = ? 
	ORDER BY trade_time DESC 
	LIMIT ?`

	rows, err := m.db.conn.Query(query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []MarketTrade
	for rows.Next() {
		var t MarketTrade
		err := rows.Scan(&t.Symbol, &t.Price, &t.Quantity, &t.TradeTime, &t.IsBuyerMaker, &t.Timestamp)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, nil
}

// GetKlineData retrieves candlestick data for technical analysis
func (m *MarketDataStore) GetKlineData(symbol, interval string, limit int) ([]MarketKline, error) {
	query := `SELECT symbol, interval_type, open_price, high_price, low_price, close_price,
		volume, quote_volume, open_time, close_time, is_closed, trade_count, ts
	FROM market_klines 
	WHERE symbol = ? AND interval_type = ?
	ORDER BY open_time DESC 
	LIMIT ?`

	rows, err := m.db.conn.Query(query, symbol, interval, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []MarketKline
	for rows.Next() {
		var k MarketKline
		err := rows.Scan(&k.Symbol, &k.IntervalType, &k.OpenPrice, &k.HighPrice,
			&k.LowPrice, &k.ClosePrice, &k.Volume, &k.QuoteVolume, &k.OpenTime,
			&k.CloseTime, &k.IsClosed, &k.TradeCount, &k.Timestamp)
		if err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}
	return klines, nil
}

// GetMarketSummary retrieves latest market analysis
func (m *MarketDataStore) GetMarketSummary(symbol string) (*MarketSummary, error) {
	query := `SELECT symbol, avg_price, volume_24h, price_trend, volatility,
		support_level, resistance_level, ts
	FROM market_summary 
	WHERE symbol = ? 
	ORDER BY ts DESC 
	LIMIT 1`

	var s MarketSummary
	err := m.db.conn.QueryRow(query, symbol).Scan(
		&s.Symbol, &s.AvgPrice, &s.Volume24h, &s.PriceTrend,
		&s.Volatility, &s.SupportLevel, &s.ResistanceLevel, &s.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetTradingVolume calculates volume metrics for decision making
func (m *MarketDataStore) GetTradingVolume(symbol string, hours int) (map[string]float64, error) {
	query := `SELECT 
		COUNT(*) as trade_count,
		SUM(quantity) as total_volume,
		AVG(price) as avg_price,
		SUM(CASE WHEN is_buyer_maker = 0 THEN quantity ELSE 0 END) as buy_volume,
		SUM(CASE WHEN is_buyer_maker = 1 THEN quantity ELSE 0 END) as sell_volume
	FROM market_trades 
	WHERE symbol = ? AND trade_time >= DATE_SUB(NOW(), INTERVAL ? HOUR)`

	var tradeCount, totalVolume, avgPrice, buyVolume, sellVolume sql.NullFloat64
	err := m.db.conn.QueryRow(query, symbol, hours).Scan(
		&tradeCount, &totalVolume, &avgPrice, &buyVolume, &sellVolume,
	)
	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"trade_count":  tradeCount.Float64,
		"total_volume": totalVolume.Float64,
		"avg_price":    avgPrice.Float64,
		"buy_volume":   buyVolume.Float64,
		"sell_volume":  sellVolume.Float64,
		"volume_ratio": buyVolume.Float64 / (buyVolume.Float64 + sellVolume.Float64),
	}, nil
}

// GetAdvancedSignals uses TiDB's analytical capabilities for sophisticated trading signals
func (m *MarketDataStore) GetAdvancedSignals(symbol string) (map[string]interface{}, error) {
	// TiDB Time-Series Analysis with Window Functions
	query := `
	WITH price_analysis AS (
		SELECT 
			price,
			ts,
			LAG(price, 1) OVER (ORDER BY ts) as prev_price,
			LAG(price, 5) OVER (ORDER BY ts) as price_5min_ago,
			LAG(price, 15) OVER (ORDER BY ts) as price_15min_ago,
			AVG(price) OVER (ORDER BY ts ROWS BETWEEN 9 PRECEDING AND CURRENT ROW) as sma_10,
			AVG(price) OVER (ORDER BY ts ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as sma_20,
			STDDEV(price) OVER (ORDER BY ts ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as volatility,
			ROW_NUMBER() OVER (ORDER BY ts DESC) as rn
		FROM market_prices 
		WHERE symbol = ? AND ts >= DATE_SUB(NOW(), INTERVAL 2 HOUR)
		ORDER BY ts DESC
	),
	volume_analysis AS (
		SELECT 
			SUM(CASE WHEN is_buyer_maker = 0 THEN quantity ELSE 0 END) as recent_buy_volume,
			SUM(CASE WHEN is_buyer_maker = 1 THEN quantity ELSE 0 END) as recent_sell_volume,
			COUNT(*) as trade_frequency,
			AVG(price) as avg_trade_price
		FROM market_trades 
		WHERE symbol = ? AND trade_time >= DATE_SUB(NOW(), INTERVAL 30 MINUTE)
	),
	support_resistance AS (
		SELECT 
			MIN(price) as support_level,
			MAX(price) as resistance_level,
			AVG(price) as mid_point
		FROM market_prices 
		WHERE symbol = ? AND ts >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
	)
	SELECT 
		p.price as current_price,
		p.prev_price,
		p.price_5min_ago,
		p.price_15min_ago,
		p.sma_10,
		p.sma_20,
		p.volatility,
		v.recent_buy_volume,
		v.recent_sell_volume,
		v.trade_frequency,
		v.avg_trade_price,
		sr.support_level,
		sr.resistance_level,
		sr.mid_point
	FROM price_analysis p, volume_analysis v, support_resistance sr
	WHERE p.rn = 1`

	var currentPrice, prevPrice, price5min, price15min, sma10, sma20, volatility sql.NullFloat64
	var buyVol, sellVol, tradeFreq, avgTradePrice, support, resistance, midPoint sql.NullFloat64

	err := m.db.conn.QueryRow(query, symbol, symbol, symbol).Scan(
		&currentPrice, &prevPrice, &price5min, &price15min, &sma10, &sma20, &volatility,
		&buyVol, &sellVol, &tradeFreq, &avgTradePrice, &support, &resistance, &midPoint,
	)
	if err != nil {
		return nil, err
	}

	// Calculate advanced indicators
	result := map[string]interface{}{
		"current_price":    currentPrice.Float64,
		"previous_price":   prevPrice.Float64,
		"price_5min_ago":   price5min.Float64,
		"price_15min_ago":  price15min.Float64,
		"sma_10":           sma10.Float64,
		"sma_20":           sma20.Float64,
		"volatility":       volatility.Float64,
		"buy_volume":       buyVol.Float64,
		"sell_volume":      sellVol.Float64,
		"trade_frequency":  tradeFreq.Float64,
		"avg_trade_price":  avgTradePrice.Float64,
		"support_level":    support.Float64,
		"resistance_level": resistance.Float64,
		"mid_point":        midPoint.Float64,
	}

	// Calculate derived signals
	if currentPrice.Float64 > 0 && sma10.Float64 > 0 && sma20.Float64 > 0 {
		result["price_vs_sma10"] = (currentPrice.Float64 - sma10.Float64) / sma10.Float64 * 100
		result["price_vs_sma20"] = (currentPrice.Float64 - sma20.Float64) / sma20.Float64 * 100
		result["sma_cross"] = sma10.Float64 > sma20.Float64 // Golden cross indicator
	}

	if buyVol.Float64+sellVol.Float64 > 0 {
		result["volume_ratio"] = buyVol.Float64 / (buyVol.Float64 + sellVol.Float64)
	}

	if currentPrice.Float64 > 0 && support.Float64 > 0 && resistance.Float64 > 0 {
		result["support_distance"] = (currentPrice.Float64 - support.Float64) / support.Float64 * 100
		result["resistance_distance"] = (resistance.Float64 - currentPrice.Float64) / currentPrice.Float64 * 100
	}

	if prevPrice.Float64 > 0 {
		result["momentum_1min"] = (currentPrice.Float64 - prevPrice.Float64) / prevPrice.Float64 * 100
	}
	if price5min.Float64 > 0 {
		result["momentum_5min"] = (currentPrice.Float64 - price5min.Float64) / price5min.Float64 * 100
	}
	if price15min.Float64 > 0 {
		result["momentum_15min"] = (currentPrice.Float64 - price15min.Float64) / price15min.Float64 * 100
	}

	return result, nil
}

// GetRealTimeMarketState uses TiDB's real-time capabilities for instant analysis
func (m *MarketDataStore) GetRealTimeMarketState(symbol string) (map[string]interface{}, error) {
	// TiDB Real-time aggregation with TIFLASH for OLAP queries
	query := `
	SELECT 
		-- Price momentum indicators
		(SELECT price FROM market_prices WHERE symbol = ? ORDER BY ts DESC LIMIT 1) as latest_price,
		(SELECT price FROM market_prices WHERE symbol = ? ORDER BY ts DESC LIMIT 1 OFFSET 1) as prev_price,
		
		-- Volume surge detection (last 5 minutes vs previous 5 minutes)
		(SELECT COALESCE(SUM(quantity), 0) FROM market_trades 
		 WHERE symbol = ? AND trade_time >= DATE_SUB(NOW(), INTERVAL 5 MINUTE)) as volume_last_5min,
		(SELECT COALESCE(SUM(quantity), 0) FROM market_trades 
		 WHERE symbol = ? AND trade_time BETWEEN DATE_SUB(NOW(), INTERVAL 10 MINUTE) 
		 AND DATE_SUB(NOW(), INTERVAL 5 MINUTE)) as volume_prev_5min,
		
		-- Order book pressure (from recent trades)
		(SELECT COUNT(*) FROM market_trades 
		 WHERE symbol = ? AND is_buyer_maker = 0 AND trade_time >= DATE_SUB(NOW(), INTERVAL 2 MINUTE)) as recent_buys,
		(SELECT COUNT(*) FROM market_trades 
		 WHERE symbol = ? AND is_buyer_maker = 1 AND trade_time >= DATE_SUB(NOW(), INTERVAL 2 MINUTE)) as recent_sells,
		
		-- Volatility spike detection
		(SELECT STDDEV(price) FROM market_prices 
		 WHERE symbol = ? AND ts >= DATE_SUB(NOW(), INTERVAL 15 MINUTE)) as volatility_15min,
		(SELECT STDDEV(price) FROM market_prices 
		 WHERE symbol = ? AND ts >= DATE_SUB(NOW(), INTERVAL 1 HOUR)) as volatility_1hour
	`

	var latestPrice, prevPrice, vol5min, volPrev5min sql.NullFloat64
	var recentBuys, recentSells sql.NullInt64
	var vol15min, vol1hour sql.NullFloat64

	err := m.db.conn.QueryRow(query, symbol, symbol, symbol, symbol, symbol, symbol, symbol, symbol).Scan(
		&latestPrice, &prevPrice, &vol5min, &volPrev5min, &recentBuys, &recentSells, &vol15min, &vol1hour,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"latest_price":     latestPrice.Float64,
		"previous_price":   prevPrice.Float64,
		"volume_last_5min": vol5min.Float64,
		"volume_prev_5min": volPrev5min.Float64,
		"recent_buys":      recentBuys.Int64,
		"recent_sells":     recentSells.Int64,
		"volatility_15min": vol15min.Float64,
		"volatility_1hour": vol1hour.Float64,
	}

	// Calculate real-time signals
	if prevPrice.Float64 > 0 {
		result["price_momentum"] = (latestPrice.Float64 - prevPrice.Float64) / prevPrice.Float64 * 100
	}

	if volPrev5min.Float64 > 0 {
		result["volume_surge"] = vol5min.Float64 / volPrev5min.Float64
	}

	if recentBuys.Int64+recentSells.Int64 > 0 {
		result["buy_pressure"] = float64(recentBuys.Int64) / float64(recentBuys.Int64+recentSells.Int64)
	}

	if vol1hour.Float64 > 0 {
		result["volatility_spike"] = vol15min.Float64 / vol1hour.Float64
	}

	return result, nil
}
